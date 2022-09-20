package attendeesrv

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
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

func (s *AttendeeServiceImplData) DoStatusChange(ctx context.Context, attendee *entity.Attendee, newStatus string, comments string) error {
	// controller checks value validity
	// controller checks permission via StatusChangeAllowed
	// controller checks precondition via StatusChangePossible

	if newStatus == "approved" {
		// the other dues updates come during attendee updates with package changes
		err := s.UpdateDues(ctx, attendee)
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

	// TODO: add sending notification emails here

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

	switch newStatus {
	case "new":
		return s.checkZeroOrNegativePaymentBalance(ctx, attendee)
	case "approved":
		return s.checkZeroOrNegativePaymentBalance(ctx, attendee)
	case "partially paid":
		return s.checkPositivePaymentBalance(ctx, attendee)
	case "paid":
		return s.checkPaidInFullWithGraceAmount(ctx, attendee)
	case "checked in":
		return s.checkPaidInFull(ctx, attendee)
	case "cancelled":
		return nil
	case "deleted":
		return s.checkNoPaymentsExist(ctx, attendee)
	default:
		return UnknownStatusError
	}
}

func (s *AttendeeServiceImplData) checkNoPaymentsExist(ctx context.Context, attendee *entity.Attendee) error {
	// TODO implement me
	return CannotDeleteError
}

func (s *AttendeeServiceImplData) checkZeroOrNegativePaymentBalance(ctx context.Context, attendee *entity.Attendee) error {
	// TODO implement me
	return HasPaymentBalanceError
}

func (s *AttendeeServiceImplData) checkPositivePaymentBalance(ctx context.Context, attendee *entity.Attendee) error {
	// TODO implement me
	return InsufficientPaymentError
}

func (s *AttendeeServiceImplData) checkPaidInFullWithGraceAmount(ctx context.Context, attendee *entity.Attendee) error {
	// TODO implement me - attn, a guest may have 0 balance
	return InsufficientPaymentError
}

func (s *AttendeeServiceImplData) checkPaidInFull(ctx context.Context, attendee *entity.Attendee) error {
	// TODO implement me - attn, a guest may have 0 balance
	return InsufficientPaymentError
}
