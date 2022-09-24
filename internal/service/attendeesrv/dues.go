package attendeesrv

import (
	"context"
	"errors"
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"strconv"
	"time"
)

func (s *AttendeeServiceImplData) UpdateDues(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) (string, error) {
	transactionHistory, err := paymentservice.Get().GetTransactions(ctx, attendee.ID)
	if err != nil && !errors.Is(err, paymentservice.NoSuchDebitor404Error) {
		return newStatus, err
	}

	if newStatus == "new" || newStatus == "deleted" {
		err = s.compensateAllDues(ctx, attendee, newStatus, transactionHistory)
		if err != nil {
			return newStatus, err
		}
	} else if newStatus == "cancelled" {
		err = s.compensateUnpaidDuesOnCancel(ctx, attendee, transactionHistory)
		if err != nil {
			return newStatus, err
		}
	} else {
		err = s.adjustDuesAccordingToSelectedPackages(ctx, attendee, transactionHistory)
		if err != nil {
			return newStatus, err
		}

		if newStatus == "approved" || newStatus == "partially paid" || newStatus == "paid" {
			// we do not adjust status back once checked in

			updatedTransactionHistory, err := paymentservice.Get().GetTransactions(ctx, attendee.ID)
			if err != nil {
				return newStatus, err
			}

			// TODO status may have changed between approved <-> partially paid <-> paid
			dues, payments := s.balances(updatedTransactionHistory)

			if payments <= 0 {
				if dues > 0 {
					newStatus = "approved"
				} else {
					// guests, or has credit :)
					newStatus = "paid"
				}
			} else {
				if payments < dues {
					newStatus = "partially paid"
				} else {
					newStatus = "paid"
				}
			}
		}
	}

	return newStatus, nil
}

func (s *AttendeeServiceImplData) adjustDuesAccordingToSelectedPackages(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	oldDuesByVAT := s.oldDuesByVAT(transactionHistory)
	packageDuesByVAT := s.packageDuesByVAT(attendee)

	// add missing keys to packageDuesByVAT, so we can just iterate over it and not miss any tax rates
	for vatStr, _ := range oldDuesByVAT {
		_, ok := packageDuesByVAT[vatStr]
		if !ok {
			packageDuesByVAT[vatStr] = 0
		}
	}

	for vatStr, desiredBalance := range packageDuesByVAT {
		currentBalance, _ := oldDuesByVAT[vatStr]
		if currentBalance != desiredBalance {
			diffTx := s.duesTransactionForAttendee(attendee, desiredBalance-currentBalance, vatStr, "dues adjustment due to change in status or selected packages")
			err := paymentservice.Get().AddTransaction(ctx, diffTx)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *AttendeeServiceImplData) packageDuesByVAT(attendee *entity.Attendee) map[string]int64 {
	result := make(map[string]int64)
	packageConfigs := config.Configuration().Choices.Packages
	for key, selected := range choiceStrToMap(attendee.Packages) {
		if selected {
			packageConfig, ok := packageConfigs[key]
			if !ok {
				// TODO attendee has package that is not configured - log as error and discard
			} else {
				vatStr := fmt.Sprintf("%.6f", packageConfig.VatPercent)

				// TODO IMPORTANT determine whether to use early, late, or atcon dues rate, based on time constraints in config
				price := int64(packageConfig.PriceEarly * 100)

				previous, _ := result[vatStr]
				result[vatStr] = previous + price
			}
		}
	}
	return result
}

func (s *AttendeeServiceImplData) compensateAllDues(ctx context.Context, attendee *entity.Attendee, newStatus string, transactionHistory []paymentservice.Transaction) error {
	oldDuesByVAT := s.oldDuesByVAT(transactionHistory)

	// we want all dues wiped, so book negative balance for each tax rate
	comment := fmt.Sprintf("remove dues balance - status changed to %s", newStatus) // TODO language
	for vatStr, duesBalance := range oldDuesByVAT {
		compensatingTx := s.duesTransactionForAttendee(attendee, -duesBalance, vatStr, comment)
		err := paymentservice.Get().AddTransaction(ctx, compensatingTx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *AttendeeServiceImplData) compensateUnpaidDuesOnCancel(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	_, paid := s.balances(transactionHistory)
	paid += s.pseudoPaymentsFromNegativeDues(transactionHistory)

	// earliest dues get filled first
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.Type == paymentservice.Due {
			if tx.Amount.GrossCent > 0 {
				vatStr := fmt.Sprintf("%.6f", tx.Amount.VatRate)

				if paid >= tx.Amount.GrossCent {
					// the payments cover this dues transaction, keep it unchanged and reduce the available payment pool
					paid -= tx.Amount.GrossCent
				} else if paid > 0 {
					// payments partially cover the dues transaction, book compensating tx for remainder
					remainderCompensatingTx := s.duesTransactionForAttendee(attendee, tx.Amount.GrossCent-paid, vatStr, "void unpaid dues on cancel")
					err := paymentservice.Get().AddTransaction(ctx, remainderCompensatingTx)
					if err != nil {
						return err
					}
					paid = 0
				} else {
					// no payments left, compensate completely
					compensatingTx := s.duesTransactionForAttendee(attendee, tx.Amount.GrossCent, vatStr, "void unpaid dues on cancel")
					err := paymentservice.Get().AddTransaction(ctx, compensatingTx)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (s *AttendeeServiceImplData) oldDuesByVAT(transactionHistory []paymentservice.Transaction) map[string]int64 {
	oldDuesByVAT := make(map[string]int64)
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.Type == paymentservice.Due {
			vatStr := fmt.Sprintf("%.6f", tx.Amount.VatRate)

			// TODO support multiple currencies, or at least read currency from config and reject any not in this currency

			previous, _ := oldDuesByVAT[vatStr]
			oldDuesByVAT[vatStr] = previous + tx.Amount.GrossCent
		}
	}
	return oldDuesByVAT
}

func (s *AttendeeServiceImplData) balances(transactionHistory []paymentservice.Transaction) (validDues int64, validPayments int64) {
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid {
			if tx.Type == paymentservice.Payment {
				validPayments += tx.Amount.GrossCent
			} else if tx.Type == paymentservice.Due {
				validDues += tx.Amount.GrossCent
			}
		}
	}
	return
}

func (s *AttendeeServiceImplData) pseudoPaymentsFromNegativeDues(transactionHistory []paymentservice.Transaction) (validNegativeDuesSum int64) {
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.Type == paymentservice.Due {
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
		DebitorID: attendee.ID,
		Type:      paymentservice.Due,
		Method:    paymentservice.Internal,
		Amount: paymentservice.Amount{
			Currency:  "EUR", // TODO from config
			GrossCent: amount,
			VatRate:   vat,
		},
		Comment:       comment,
		Status:        paymentservice.Valid,
		EffectiveDate: "",          // TODO - dues are effective immediately
		DueDate:       time.Time{}, // TODO - implement weeks logic, except for negative amounts
	}
}
