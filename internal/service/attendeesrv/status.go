package attendeesrv

import (
	"context"
	"errors"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/mailservice"
	"github.com/eurofurence/reg-attendee-service/internal/repository/paymentservice"
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
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

func (s *AttendeeServiceImplData) UpdateDuesAndDoStatusChangeIfNeeded(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string, comments string) error {
	var err error
	// controller checks value validity
	// controller checks permission via StatusChangeAllowed
	// controller checks precondition via StatusChangePossible

	// Note that UpdateDues may adjust the status according to payment balance
	newStatus, err = s.UpdateDues(ctx, attendee, oldStatus, newStatus)
	if err != nil {
		return err
	}

	if newStatus != oldStatus {
		change := entity.StatusChange{
			AttendeeId: attendee.ID,
			Status:     newStatus,
			Comments:   comments,
		}
		err = database.GetRepository().AddStatusChange(ctx, &change)
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
	}

	return nil
}

func (s *AttendeeServiceImplData) StatusChangeAllowed(ctx context.Context, attendee *entity.Attendee, oldStatus string, newStatus string) error {
	if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsRole(ctx, config.OidcAdminRole()) {
		// api or admin
		return nil
	}

	subject := ctxvalues.Subject(ctx)
	if subject == "" {
		// anon
		return errors.New("all status changes require a logged in user")
	}

	if subject == attendee.Identity {
		// self
		if oldStatus == "new" && newStatus == "cancelled" || oldStatus == "approved" && newStatus == "cancelled" {
			aulogging.Logger.Ctx(ctx).Info().Printf("self cancellation for attendee %d by %s", attendee.ID, subject)
			return nil
		}

		aulogging.Logger.Ctx(ctx).Warn().Printf("forbidden self status change attempt %s -> %s for attendee %d by %s", oldStatus, newStatus, attendee.ID, subject)
		return errors.New("you are not allowed to make this status transition - the attempt has been logged")
	}

	// others

	if oldStatus == "paid" && newStatus == "checked in" {
		// TODO - this is kind of ugly

		// check that any of the registrations owned by subject have the regdesk permission
		ownedAttendees, err := database.GetRepository().FindByIdentity(ctx, subject)
		if err != nil {
			return err
		}
		for _, oa := range ownedAttendees {
			adminInfo, err := database.GetRepository().GetAdminInfoByAttendeeId(ctx, oa.ID)
			if err != nil {
				return err
			}
			permissions := choiceStrToMap(adminInfo.Permissions)
			allowed, _ := permissions["regdesk"]
			if allowed {
				aulogging.Logger.Ctx(ctx).Info().Printf("regdesk check in for attendee %d by %s", attendee.ID, subject)
				return nil
			}
		}
	}

	aulogging.Logger.Ctx(ctx).Warn().Printf("forbidden status change attempt %s -> %s for attendee %d by %s", oldStatus, newStatus, attendee.ID, subject)
	return errors.New("you are not allowed to make this status transition - the attempt has been logged")
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
		if oldStatus == "new" || oldStatus == "cancelled" || oldStatus == "deleted" {
			return GoToApprovedFirst
		}
		return s.checkPositivePaymentBalanceButNotFullPayment(ctx, attendee, transactionHistory)
	case "paid":
		if oldStatus == "new" || oldStatus == "cancelled" || oldStatus == "deleted" {
			return GoToApprovedFirst
		}
		return s.checkPaidInFullWithGraceAmount(ctx, attendee, transactionHistory)
	case "checked in":
		if oldStatus == "new" || oldStatus == "cancelled" || oldStatus == "deleted" {
			return GoToApprovedFirst
		}
		return s.checkPaidInFull(ctx, attendee, transactionHistory)
	case "cancelled":
		return nil
	case "deleted":
		return s.checkNoPaymentsExist(ctx, attendee, transactionHistory)
	default:
		return UnknownStatusError
	}
}

var graceAmountCents int64 = 100 // TODO read from config

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
	if paid >= dues-graceAmountCents {
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

func (s *AttendeeServiceImplData) IsOwnerFor(ctx context.Context) ([]*entity.Attendee, error) {
	identity := ctxvalues.Subject(ctx)
	if identity != "" {
		return database.GetRepository().FindByIdentity(ctx, identity)
	} else {
		return make([]*entity.Attendee, 0), nil
	}
}
