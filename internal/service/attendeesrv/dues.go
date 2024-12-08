package attendeesrv

import (
	"context"
	"errors"
	"fmt"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"strconv"
	"strings"
)

func (s *AttendeeServiceImplData) UpdateDuesTransactions(ctx context.Context, attendee *entity.Attendee, newStatus status.Status, commentOverride string) ([]paymentservice.Transaction, *entity.AdminInfo, error) {
	adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, attendee.ID)
	if err != nil {
		return []paymentservice.Transaction{}, adminInfo, err
	}

	transactionHistory, err := paymentservice.Get().GetTransactions(ctx, attendee.ID)
	if err != nil && !errors.Is(err, paymentservice.NoSuchDebitor404Error) {
		return []paymentservice.Transaction{}, adminInfo, err
	}

	updated := false
	if newStatus == status.New || newStatus == status.Deleted || newStatus == status.Waiting {
		updated, err = s.compensateAllDues(ctx, attendee, newStatus, transactionHistory)
		if err != nil {
			return transactionHistory, adminInfo, err
		}
	} else if newStatus == status.Cancelled {
		updated, err = s.compensateUnpaidDuesOnCancel(ctx, attendee, transactionHistory)
		if err != nil {
			return transactionHistory, adminInfo, err
		}
	} else {
		updated, err = s.adjustDuesAccordingToSelectedPackages(ctx, attendee, adminInfo, transactionHistory, commentOverride)
		if err != nil {
			return transactionHistory, adminInfo, err
		}
	}

	updatedTransactionHistory := transactionHistory
	if updated {
		updatedTransactionHistory, err = paymentservice.Get().GetTransactions(ctx, attendee.ID)
		if err != nil && !errors.Is(err, paymentservice.NoSuchDebitor404Error) {
			return []paymentservice.Transaction{}, adminInfo, err
		}
	}

	return updatedTransactionHistory, adminInfo, nil
}

func (s *AttendeeServiceImplData) compensateAllDues(ctx context.Context, attendee *entity.Attendee, newStatus status.Status, transactionHistory []paymentservice.Transaction) (bool, error) {
	oldDuesByVAT := s.oldDuesByVAT(transactionHistory)
	updated := false

	// we want all dues wiped, so book negative balance for each tax rate
	comment := fmt.Sprintf("remove dues balance - status changed to %s", newStatus) // TODO language
	for vatStr, duesBalance := range oldDuesByVAT {
		if duesBalance != 0 {
			updated = true
			compensatingTx := s.duesTransactionForAttendee(attendee, -duesBalance, vatStr, comment)
			err := paymentservice.Get().AddTransaction(ctx, compensatingTx)
			if err != nil {
				return updated, err
			}
		}
	}
	return updated, nil
}

func (s *AttendeeServiceImplData) compensateUnpaidDuesOnCancel(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) (bool, error) {
	_, paid, _, _ := s.balances(transactionHistory)
	paid += s.pseudoPaymentsFromNegativeDues(transactionHistory)
	updated := false

	// earliest dues get filled first
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.TransactionType == paymentservice.Due {
			if tx.Amount.GrossCent > 0 {
				updated = true
				vatStr := fmt.Sprintf("%.6f", tx.Amount.VatRate)

				if paid >= tx.Amount.GrossCent {
					// the payments cover this dues transaction, keep it unchanged and reduce the available payment pool
					paid -= tx.Amount.GrossCent
				} else if paid > 0 {
					// payments partially cover the dues transaction, book compensating tx for remainder
					remainderCompensatingTx := s.duesTransactionForAttendee(attendee, -(tx.Amount.GrossCent - paid), vatStr, "void unpaid dues on cancel")
					err := paymentservice.Get().AddTransaction(ctx, remainderCompensatingTx)
					if err != nil {
						return updated, err
					}
					paid = 0
				} else {
					// no payments left, compensate completely
					compensatingTx := s.duesTransactionForAttendee(attendee, -tx.Amount.GrossCent, vatStr, "void unpaid dues on cancel")
					err := paymentservice.Get().AddTransaction(ctx, compensatingTx)
					if err != nil {
						return updated, err
					}
				}
			}
		}
	}
	return updated, nil
}

func (s *AttendeeServiceImplData) adjustDuesAccordingToSelectedPackages(ctx context.Context, attendee *entity.Attendee, adminInfo *entity.AdminInfo, transactionHistory []paymentservice.Transaction, commentOverride string) (bool, error) {
	oldDuesByVAT := s.oldDuesByVAT(transactionHistory)
	packageDuesByVAT := s.packageDuesByVAT(ctx, attendee, adminInfo)
	updated := false

	// add missing keys to packageDuesByVAT, so we can just iterate over it and not miss any tax rates
	for vatStr, _ := range oldDuesByVAT {
		_, ok := packageDuesByVAT[vatStr]
		if !ok {
			packageDuesByVAT[vatStr] = 0
		}
	}

	comment := "dues adjustment due to change in status or selected packages"
	if commentOverride != "" {
		comment = commentOverride
	}

	for vatStr, desiredBalance := range packageDuesByVAT {
		currentBalance, _ := oldDuesByVAT[vatStr]
		if currentBalance != desiredBalance {
			updated = true
			diffTx := s.duesTransactionForAttendee(attendee, desiredBalance-currentBalance, vatStr, comment)
			err := paymentservice.Get().AddTransaction(ctx, diffTx)
			if err != nil {
				return updated, err
			}
		}
	}

	return updated, nil
}

func (s *AttendeeServiceImplData) packageDuesByVAT(ctx context.Context, attendee *entity.Attendee, adminInfo *entity.AdminInfo) map[string]int64 {
	result := make(map[string]int64)

	// consider manual dues before guest status (they might be due a refund from last year, or something)
	if adminInfo.ManualDues != 0 {
		vatStr := fmt.Sprintf("%.6f", config.VatPercent())
		result[vatStr] = adminInfo.ManualDues
	}

	if s.considerGuest(ctx, adminInfo) {
		// guests pay nothing for ANY normal packages
		return result
	}

	packageConfigs := config.PackagesConfig()
	for key, count := range choiceStrToMap(attendee.Packages, packageConfigs) {
		if count > 0 {
			packageConfig, ok := packageConfigs[key]
			if !ok {
				aulogging.Logger.Ctx(ctx).Warn().Printf("attendee id %d has unknown package %s in db - ignoring during dues calculation", attendee.ID, key)
			} else {
				vatStr := fmt.Sprintf("%.6f", packageConfig.VatPercent)

				previous, _ := result[vatStr]
				result[vatStr] = previous + packageConfig.Price*int64(count)
			}
		}
	}
	return result
}

func (s *AttendeeServiceImplData) considerGuest(ctx context.Context, adminInfo *entity.AdminInfo) bool {
	adminFlagsMap := choiceStrToMap(adminInfo.Flags, config.FlagsConfigAdminOnly())
	isGuest, ok := adminFlagsMap["guest"]
	if !ok {
		aulogging.Logger.Ctx(ctx).Warn().Print("admin only flag 'guest' not configured, skipping")
	}
	return isGuest > 0
}

func (s *AttendeeServiceImplData) oldDuesByVAT(transactionHistory []paymentservice.Transaction) map[string]int64 {
	oldDuesByVAT := make(map[string]int64)
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.TransactionType == paymentservice.Due {
			vatStr := fmt.Sprintf("%.6f", tx.Amount.VatRate)

			previous, _ := oldDuesByVAT[vatStr]
			oldDuesByVAT[vatStr] = previous + tx.Amount.GrossCent
		}
	}
	return oldDuesByVAT
}

// ---

func (s *AttendeeServiceImplData) UpdateAttendeeCacheAndCalculateResultingStatus(ctx context.Context, attendee *entity.Attendee, updatedTransactionHistory []paymentservice.Transaction, newStatus status.Status) (status.Status, bool, error) {
	// identity and zip each get an id dependent suffix for deleted attendees to ensure the user can register again after deletion
	// (identity has a unique index in the db!)
	// (nick, email, zip has a unique index in the db!)
	identity := s.suffixForDeletedAttendees(attendee, newStatus, attendee.Identity)
	zip := s.suffixForDeletedAttendees(attendee, newStatus, attendee.Zip)

	dues, payments, open, dueDate := s.balances(updatedTransactionHistory)
	// never move due date back in time (allows manual override)
	if attendee.CacheDueDate != "" && attendee.CacheDueDate > dueDate {
		dueDate = attendee.CacheDueDate
	}

	duesInformationChanged, err := s.updateCachedValuesAndIdentityInAttendee(ctx, attendee, dues, payments, open, dueDate, identity, zip)
	if err != nil {
		return newStatus, false, err
	}

	if newStatus == status.Approved || newStatus == status.PartiallyPaid || newStatus == status.Paid {
		// we do not adjust status back once checked in
		newStatus = s.calculateResultingStatusForApprovedToPaid(payments, dues)
	}

	return newStatus, duesInformationChanged, nil
}

func (s *AttendeeServiceImplData) suffixForDeletedAttendees(attendee *entity.Attendee, newStatus status.Status, value string) string {
	deletionSuffix := fmt.Sprintf("_d_%d", attendee.ID)
	if newStatus == status.Deleted {
		if !strings.HasSuffix(value, deletionSuffix) {
			value = value + deletionSuffix
		}
	} else {
		// also prevents undelete after new registration has been made
		value = strings.TrimSuffix(value, deletionSuffix)
	}
	return value
}

func (s *AttendeeServiceImplData) updateCachedValuesAndIdentityInAttendee(ctx context.Context, attendee *entity.Attendee, dues int64, payments int64, open int64, dueDate string, identity string, zip string) (bool, error) {
	duesRelevantUpdate := attendee.CacheTotalDues != dues ||
		attendee.CachePaymentBalance != payments ||
		attendee.CacheDueDate != dueDate

	needsUpdate := duesRelevantUpdate ||
		attendee.CacheOpenBalance != open ||
		attendee.Identity != identity ||
		attendee.Zip != zip

	if needsUpdate {
		attendee.CacheTotalDues = dues
		attendee.CachePaymentBalance = payments
		attendee.CacheOpenBalance = open
		attendee.CacheDueDate = dueDate
		attendee.Identity = identity
		attendee.Zip = zip
		err := database.GetRepository().UpdateAttendee(ctx, attendee)
		return duesRelevantUpdate, err
	}
	return duesRelevantUpdate, nil
}

func (s *AttendeeServiceImplData) calculateResultingStatusForApprovedToPaid(payments int64, dues int64) status.Status {
	if payments <= 0 {
		if dues > 0 {
			return status.Approved
		} else {
			// guests, or has credit :)
			return status.Paid
		}
	} else {
		if payments < dues-graceAmountCents {
			return status.PartiallyPaid
		} else {
			return status.Paid
		}
	}
}

func (s *AttendeeServiceImplData) balances(transactionHistory []paymentservice.Transaction) (validDues int64, validPayments int64, openPayments int64, dueDate string) {
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid {
			if tx.TransactionType == paymentservice.Payment {
				validPayments += tx.Amount.GrossCent
			} else if tx.TransactionType == paymentservice.Due {
				validDues += tx.Amount.GrossCent
				dueDate = tx.DueDate // initialize with last due date as default (most likely causes no change for non dues status values such as paid)
			}
		}
		if tx.Status == paymentservice.Tentative || tx.Status == paymentservice.Pending {
			if tx.TransactionType == paymentservice.Payment {
				openPayments += tx.Amount.GrossCent
			}
		}
	}
	dueDate = s.calculateDueDate(transactionHistory, validPayments, dueDate)
	return
}

func (s *AttendeeServiceImplData) calculateDueDate(transactionHistory []paymentservice.Transaction, validPayments int64, defaultDueDate string) string {
	var accruedDues int64
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid {
			if tx.TransactionType == paymentservice.Due {
				accruedDues += tx.Amount.GrossCent
				if accruedDues > validPayments {
					// the first incompletely paid due amount determines the due date
					return tx.DueDate
				}
			}
		}
	}
	return defaultDueDate
}

func (s *AttendeeServiceImplData) pseudoPaymentsFromNegativeDues(transactionHistory []paymentservice.Transaction) (validNegativeDuesSum int64) {
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.TransactionType == paymentservice.Due {
			if tx.Amount.GrossCent < 0 {
				// refunded tx -> count as pseudo payment
				validNegativeDuesSum += -tx.Amount.GrossCent
			}
		}
	}
	return
}

func (s *AttendeeServiceImplData) duesTransactionForAttendee(attendee *entity.Attendee, amount int64, vatStr string, comment string) paymentservice.Transaction {
	vat, _ := strconv.ParseFloat(vatStr, 64)

	return paymentservice.Transaction{
		DebitorID:       attendee.ID,
		TransactionType: paymentservice.Due,
		Method:          paymentservice.Internal,
		Amount: paymentservice.Amount{
			Currency:  config.Currency(),
			GrossCent: amount,
			VatRate:   vat,
		},
		Comment:       comment,
		Status:        paymentservice.Valid,
		EffectiveDate: s.duesEffectiveDate(),
		DueDate:       s.duesDueDate(),
	}
}

func (s *AttendeeServiceImplData) duesEffectiveDate() string {
	return s.Now().Format(config.IsoDateFormat)
}

func (s *AttendeeServiceImplData) duesDueDate() string {
	calculated := s.Now().Add(config.DueDays()).Format(config.IsoDateFormat)
	if calculated < config.EarliestDueDate() {
		return config.EarliestDueDate()
	}
	if calculated > config.LatestDueDate() {
		return config.LatestDueDate()
	}
	return calculated
}
