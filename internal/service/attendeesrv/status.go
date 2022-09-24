package attendeesrv

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/ctxvalues"
	"gorm.io/gorm"
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
		Status:     "new",
		Comments:   "registration",
	})

	for _, change := range fromDb {
		result = append(result, change)
	}

	return result, nil
}

func (s *AttendeeServiceImplData) DoStatusChange(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string, comments string) error {
	// controller checks value validity
	// controller checks permission via StatusChangeAllowed
	// controller checks precondition via StatusChangePossible

	if newStatus == "approved" {
		// the other dues updates come during attendee updates with package changes
		err := s.UpdateDues(ctx, attendee, newStatus)
		if err != nil {
			return err
		}
	}

	change := entity.StatusChange{
		AttendeeId: attendee.ID,
		Status:     newStatus,
		Comments:   comments,
	}
	err := database.GetRepository().AddStatusChange(ctx, &change)
	if err != nil {
		return err
	}

	err = mailservice.Get().SendEmail(ctx, mailservice.TemplateRequestDto{
		Name: "new-status-" + newStatus,
		Variables: map[string]string{
			"nickname": attendee.Nickname,
		},
		Email: attendee.Email,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *AttendeeServiceImplData) StatusChangeAllowed(ctx context.Context, oldStatus string, newStatus string) error {
	group, err := ctxvalues.AuthorizedAsGroup(ctx)
	if err != nil {
		return errors.New("all status changes require a logged in user")
	} else if group == config.OptionalTokenForInitialReg || group == config.TokenForLoggedInUser {
		if oldStatus == "paid" && newStatus == "checked in" {
			// TODO: load and check for regdesk permission, and allow if set
		}

		if oldStatus == "new" && newStatus == "cancelled" || oldStatus == "approved" && newStatus == "cancelled" {
			// TODO: allow for self (current model cannot check this)
		}

		return errors.New("you are not allowed to make this status transition")
	} else if group == config.TokenForAdmin {
		// admin is allowed to do all status changes
		return nil
	} else {
		return errors.New("you are not allowed to make this status transition")
	}
}

func (s *AttendeeServiceImplData) StatusChangePossible(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) error {
	if oldStatus == newStatus {
		return SameStatusError
	}

	transactionHistory, err := paymentservice.Get().GetTransactions(ctx, attendee.ID)
	if err != nil && !errors.Is(err, paymentservice.NoSuchDebitor404Error) {
		return err
	}

	switch newStatus {
	case "new":
		return s.checkZeroOrNegativePaymentBalance(ctx, attendee, transactionHistory)
	case "approved":
		return s.checkZeroOrNegativePaymentBalance(ctx, attendee, transactionHistory)
	case "partially paid":
		return s.checkPositivePaymentBalanceButNotFullPayment(ctx, attendee, transactionHistory)
	case "paid":
		return s.checkPaidInFullWithGraceAmount(ctx, attendee, transactionHistory)
	case "checked in":
		return s.checkPaidInFull(ctx, attendee, transactionHistory)
	case "cancelled":
		return nil
	case "deleted":
		return s.checkNoPaymentsExist(ctx, attendee, transactionHistory)
	default:
		return UnknownStatusError
	}
}

var graceAmount int64 = 100 // TODO read from config

func (s *AttendeeServiceImplData) checkNoPaymentsExist(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	for _, tx := range transactionHistory {
		if tx.Status == paymentservice.Valid && tx.Type == paymentservice.Payment && tx.Amount.GrossCent != 0 {
			return CannotDeleteError
		}
	}
	return nil
}

func (s *AttendeeServiceImplData) checkZeroOrNegativePaymentBalance(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	_, paid := s.balances(transactionHistory)
	if paid <= 0 {
		return nil
	} else {
		return HasPaymentBalanceError
	}
}

func (s *AttendeeServiceImplData) checkPositivePaymentBalanceButNotFullPayment(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	dues, paid := s.balances(transactionHistory)
	if paid >= 0 && paid < dues {
		return nil
	} else {
		return InsufficientPaymentError
	}
}

func (s *AttendeeServiceImplData) checkPaidInFullWithGraceAmount(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	dues, paid := s.balances(transactionHistory)
	// intentionally do not check paid >= 0, there may be negative dues (previous year refunds)
	if paid >= dues-graceAmount {
		return nil
	} else {
		return InsufficientPaymentError
	}
}

func (s *AttendeeServiceImplData) checkPaidInFull(ctx context.Context, attendee *entity.Attendee, transactionHistory []paymentservice.Transaction) error {
	dues, paid := s.balances(transactionHistory)
	if paid >= dues {
		return nil
	} else {
		return InsufficientPaymentError
	}
}
