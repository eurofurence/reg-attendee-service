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
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"gorm.io/gorm"
	"strings"
	"time"
)

func (s *AttendeeServiceImplData) GetFullStatusHistory(ctx context.Context, attendee *entity.Attendee) ([]entity.StatusChange, error) {
	// controller checks permissions

	result := make([]entity.StatusChange, 0)
	if attendee.ID == 0 {
		return result, errors.New("invalid attendee missing id, please read full dataset from the database - this is an implementation error")
	}

	fromDb, err := database.GetRepository().GetStatusChangesByAttendeeId(ctx, attendee.ID)
	if err != nil {
		return result, err
	}

	// first status entry comes from registration time, not stored in db for performance reasons during initial reg
	result = append(result, entity.StatusChange{
		Model: gorm.Model{
			CreatedAt: attendee.CreatedAt,
		},
		AttendeeId: attendee.ID,
		Status:     status.New,
		Comments:   "registration",
	})

	for _, change := range fromDb {
		result = append(result, change)
	}

	return result, nil
}

func (s *AttendeeServiceImplData) UpdateDuesAndDoStatusChangeIfNeeded(ctx context.Context, attendee *entity.Attendee, oldStatus status.Status, newStatus status.Status, statusComment string, overrideDuesComment string, suppressMinorUpdateEmail bool) error {
	// controller checks value validity
	// controller checks permission via StatusChangeAllowed
	// controller checks precondition via StatusChangePossible
	// attendee has been loaded from db in all cases

	updatedTransactionHistory, adminInfo, err := s.UpdateDuesTransactions(ctx, attendee, newStatus, overrideDuesComment)
	if err != nil {
		return err
	}

	var duesInformationChanged bool
	newStatus, duesInformationChanged, err = s.UpdateAttendeeCacheAndCalculateResultingStatus(ctx, attendee, updatedTransactionHistory, newStatus)
	if err != nil {
		return err
	}

	if newStatus != oldStatus {
		change := entity.StatusChange{
			AttendeeId: attendee.ID,
			Status:     newStatus,
			Comments:   statusComment,
		}
		err = database.GetRepository().AddStatusChange(ctx, &change)
		if err != nil {
			return err
		}

		if newStatus == status.Deleted {
			err = database.GetRepository().SoftDeleteAttendeeById(ctx, attendee.ID)
			if err != nil {
				return err
			}
		} else if oldStatus == status.Deleted {
			err = database.GetRepository().UndeleteAttendeeById(ctx, attendee.ID)
			if err != nil {
				return err
			}
		}

		if newStatus != status.Deleted && newStatus != status.CheckedIn {
			suppress := suppressMinorUpdateEmail && isPaymentPhaseStatus(oldStatus) && isPaymentPhaseStatus(newStatus)
			err = s.sendStatusChangeNotificationEmail(ctx, attendee, adminInfo, newStatus, statusComment, suppress)
			if err != nil {
				return err
			}
		}
	} else if duesInformationChanged && (newStatus == status.Approved || newStatus == status.PartiallyPaid) {
		err = s.sendStatusChangeNotificationEmail(ctx, attendee, adminInfo, newStatus, statusComment, suppressMinorUpdateEmail)
		if err != nil {
			return err
		}
	}

	return nil
}

func isPaymentPhaseStatus(st status.Status) bool {
	return st == status.Approved || st == status.PartiallyPaid || st == status.Paid
}

func formatCurr(value int64) string {
	// TODO: format currency according to provided format from config
	return fmt.Sprintf("%s %0.2f", config.Currency(), float64(value)/100.0)
}

func formatDate(value string) string {
	// TODO: format due according to provided format from config
	parsed, err := time.Parse(config.IsoDateFormat, value)
	if err != nil {
		return value
	}
	return parsed.Format(config.HumanDateFormat)
}

func (s *AttendeeServiceImplData) ResendStatusMail(ctx context.Context, attendee *entity.Attendee, currentStatus status.Status, currentStatusComment string) error {
	adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, attendee.ID)
	if err != nil {
		return err
	}

	if currentStatus != status.Deleted && currentStatus != status.CheckedIn && currentStatus != status.New {
		err = s.sendStatusChangeNotificationEmail(ctx, attendee, adminInfo, currentStatus, currentStatusComment, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *AttendeeServiceImplData) sendStatusChangeNotificationEmail(ctx context.Context, attendee *entity.Attendee, adminInfo *entity.AdminInfo, newStatus status.Status, statusComment string, suppress bool) error {
	checkSummedId := s.badgeId(attendee.ID)
	cancelReason := ""
	if newStatus == status.Cancelled {
		cancelReason = statusComment
	}
	remainingDues := attendee.CacheTotalDues - attendee.CachePaymentBalance

	dueDate := formatDate(attendee.CacheDueDate)
	if remainingDues <= 0 {
		dueDate = ""
	}

	mailDto := mailservice.MailSendDto{
		CommonID: "change-status-" + string(newStatus),
		Lang:     removeWrappingCommasWithDefault(attendee.RegistrationLanguage, "en-US"),
		Variables: map[string]string{
			"badge_number":               fmt.Sprintf("%d", attendee.ID),
			"badge_number_with_checksum": *checkSummedId,
			"nickname":                   attendee.Nickname,
			"email":                      attendee.Email,
			"reason":                     cancelReason,
			"remaining_dues":             formatCurr(remainingDues),
			"total_dues":                 formatCurr(attendee.CacheTotalDues),
			"pending_payments":           formatCurr(attendee.CacheOpenBalance),
			"due_date":                   dueDate,
			"regsys_url":                 config.RegsysPublicUrl(),

			// --- unused values ---

			// room group variables, just set so all the templates work for now
			"room_group_name":         "TODO room group name",
			"room_group_owner":        "TODO room group owner nickname",
			"room_group_owner_email":  "TODO room group owner email",
			"room_group_member":       "TODO room group member nickname",
			"room_group_member_email": "TODO room group member email",

			// other stuff that is no longer used
			"confirm_link": "TODO confirmation link",
			"new_email":    "TODO email change new email",
		},
		To: []string{attendee.Email},
	}

	if s.considerGuest(ctx, adminInfo) {
		if newStatus == status.Approved || newStatus == status.PartiallyPaid || newStatus == status.Paid {
			mailDto.CommonID = "guest"
		}
	} else if suppress {
		aulogging.Logger.Ctx(ctx).Info().Printf("sending mail %s to %s suppressed", mailDto.CommonID, attendee.Email)
		return nil
	}

	err := mailservice.Get().SendEmail(ctx, mailDto)
	if err != nil {
		return err
	}
	return nil
}

func removeWrappingCommasWithDefault(v string, defaultValue string) string {
	v = strings.TrimPrefix(v, ",")
	v = strings.TrimSuffix(v, ",")
	if v == "" {
		return defaultValue
	}
	return v
}

func (s *AttendeeServiceImplData) StatusChangeAllowed(ctx context.Context, attendee *entity.Attendee, oldStatus status.Status, newStatus status.Status) error {
	if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup()) {
		// api or admin
		return nil
	}

	subject := ctxvalues.Subject(ctx)
	if subject == "" {
		// anon
		return errors.New("all status changes require a logged in user")
	}

	if subject == attendee.Identity {
		// self cancellation
		if newStatus == status.Cancelled {
			if oldStatus == status.New || oldStatus == status.Approved || oldStatus == status.Waiting {
				aulogging.Logger.Ctx(ctx).Info().Printf("self cancellation for attendee %d by %s", attendee.ID, subject)
				return nil
			}
		}

		aulogging.Logger.Ctx(ctx).Warn().Printf("forbidden self status change attempt %s -> %s for attendee %d by %s", oldStatus, newStatus, attendee.ID, subject)
		return errors.New("you are not allowed to make this status transition - the attempt has been logged")
	}

	// others

	if oldStatus == status.Paid && newStatus == status.CheckedIn {
		allowed, err := s.subjectHasAreaPermissionEntry(ctx, subject, "regdesk")
		if err != nil {
			return err
		}
		if allowed {
			aulogging.Logger.Ctx(ctx).Info().Printf("regdesk check in for attendee %d by %s", attendee.ID, subject)
			return nil
		}
	}

	aulogging.Logger.Ctx(ctx).Warn().Printf("forbidden status change attempt %s -> %s for attendee %d by %s", oldStatus, newStatus, attendee.ID, subject)
	return errors.New("you are not allowed to make this status transition - the attempt has been logged")
}

func (s *AttendeeServiceImplData) StatusChangePossible(ctx context.Context, attendee *entity.Attendee, oldStatus status.Status, newStatus status.Status) error {
	if oldStatus == newStatus {
		return SameStatusError
	}

	transactionHistory, err := paymentservice.Get().GetTransactions(ctx, attendee.ID)
	if err != nil && !errors.Is(err, paymentservice.NoSuchDebitor404Error) {
		return err
	}

	switch newStatus {
	case status.New:
		return s.checkZeroOrNegativePaymentBalance(ctx, attendee, transactionHistory)
	case status.Waiting:
		return s.checkZeroOrNegativePaymentBalance(ctx, attendee, transactionHistory)
	case status.Approved:
		if oldStatus == status.New || oldStatus == status.Waiting || oldStatus == status.Cancelled || oldStatus == status.Deleted {
			if err := s.matchesBanAndNoSkip(ctx, attendee); err != nil {
				return err
			}
		}
		return nil // explicitly allow "approved" for people with a payment balance (auto-skips ahead to partially paid or paid)
	case status.PartiallyPaid:
		if oldStatus == status.New || oldStatus == status.Waiting || oldStatus == status.Cancelled || oldStatus == status.Deleted {
			return GoToApprovedFirst
		}
		return s.checkPositivePaymentBalanceButNotFullPayment(ctx, attendee, transactionHistory)
	case status.Paid:
		if oldStatus == status.New || oldStatus == status.Waiting || oldStatus == status.Cancelled || oldStatus == status.Deleted {
			return GoToApprovedFirst
		}
		return s.checkPaidInFullWithGraceAmount(ctx, attendee, transactionHistory)
	case status.CheckedIn:
		if oldStatus == status.New || oldStatus == status.Waiting || oldStatus == status.Cancelled || oldStatus == status.Deleted {
			return GoToApprovedFirst
		}
		return s.checkPaidInFull(ctx, attendee, transactionHistory)
	case status.Cancelled:
		return nil
	case status.Deleted:
		return s.checkNoPaymentsExist(ctx, attendee, transactionHistory)
	default:
		return UnknownStatusError
	}
}

var graceAmountCents int64 = 100 // TODO read from config

func (s *AttendeeServiceImplData) checkNoPaymentsExist(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.TransactionType == paymentservice.Payment && tx.Amount.GrossCent != 0 {
			return CannotDeleteError
		}
	}
	return nil
}

func (s *AttendeeServiceImplData) checkZeroOrNegativePaymentBalance(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	_, paid, _, _ := s.balances(transactionHistory)
	if paid <= 0 {
		return nil
	} else {
		return HasPaymentBalanceError
	}
}

func (s *AttendeeServiceImplData) checkPositivePaymentBalanceButNotFullPayment(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	dues, paid, _, _ := s.balances(transactionHistory)
	if paid >= 0 && paid < dues {
		return nil
	} else {
		return InsufficientPaymentError
	}
}

func (s *AttendeeServiceImplData) checkPaidInFullWithGraceAmount(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	dues, paid, _, _ := s.balances(transactionHistory)
	// intentionally do not check paid >= 0, there may be negative dues (previous year refunds)
	if paid >= dues-graceAmountCents {
		return nil
	} else {
		return InsufficientPaymentError
	}
}

func (s *AttendeeServiceImplData) checkPaidInFull(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	dues, paid, _, _ := s.balances(transactionHistory)
	if paid >= dues {
		return nil
	} else {
		return InsufficientPaymentError
	}
}

func (s *AttendeeServiceImplData) IsOwnerFor(ctx context.Context) ([]*entity.Attendee, error) {
	identity := ctxvalues.Subject(ctx)
	if identity != "" {
		return database.GetRepository().FindByIdentity(ctx, identity)
	} else {
		return make([]*entity.Attendee, 0), nil
	}
}
