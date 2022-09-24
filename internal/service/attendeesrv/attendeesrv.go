package attendeesrv

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/filter/ctxvalues"
	"strings"
	"time"
)

func (s *AttendeeServiceImplData) NewAttendee(ctx context.Context) *entity.Attendee {
	return &entity.Attendee{}
}

func (s *AttendeeServiceImplData) RegisterNewAttendee(ctx context.Context, attendee *entity.Attendee) (uint, error) {
	alreadyExists, err := isDuplicateAttendee(ctx, attendee.Nickname, attendee.Zip, attendee.Email, 0)
	if err != nil {
		return 0, err
	}
	if alreadyExists {
		logging.Ctx(ctx).Warnf("received new registration duplicate - nick %v zip %v email %v", attendee.Nickname, attendee.Zip, attendee.Email)
		return 0, errors.New("duplicate attendee data - you are already registered")
	}

	id, err := database.GetRepository().AddAttendee(ctx, attendee)
	return id, err
}

func (s *AttendeeServiceImplData) GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error) {
	attendee, err := database.GetRepository().GetAttendeeById(ctx, id)
	return attendee, err
}

func (s *AttendeeServiceImplData) UpdateAttendee(ctx context.Context, attendee *entity.Attendee) error {
	alreadyExists, err := isDuplicateAttendee(ctx, attendee.Nickname, attendee.Zip, attendee.Email, 1)
	if err != nil {
		return err
	}
	if alreadyExists {
		logging.Ctx(ctx).Warnf("received update with registration duplicate - nick %v zip %v email %v", attendee.Nickname, attendee.Zip, attendee.Email)
		return errors.New("your changes would lead to duplicate attendee data - same nickname, zip, email")
	}

	// TODO: verify permissions - after first payment, only admins can remove packages

	err = database.GetRepository().UpdateAttendee(ctx, attendee)
	if err != nil {
		return err
	}

	statusHistory, err := s.GetFullStatusHistory(ctx, attendee)
	if err != nil {
		return err
	}

	currentStatus := statusHistory[len(statusHistory)-1].Status

	// TODO record who made the change in comment
	err = s.UpdateDuesAndDoStatusChangeIfNeeded(ctx, attendee, currentStatus, currentStatus, "attendee update")
	if err != nil {
		return err
	}

	return nil
}

func (s *AttendeeServiceImplData) GetAttendeeMaxId(ctx context.Context) (uint, error) {
	max, err := database.GetRepository().MaxAttendeeId(ctx)
	return max, err
}

func (s *AttendeeServiceImplData) CanChangeChoiceTo(ctx context.Context, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig) error {
	originalChoices := choiceStrToMap(originalChoiceStr)
	newChoices := choiceStrToMap(newChoiceStr)
	for k, v := range configuration {
		if err := checkNoForbiddenChanges(ctx, k, v, originalChoices, newChoices); err != nil {
			return err
		}
		if err := checkNoConstraintViolation(k, v, newChoices); err != nil {
			return err
		}
	}
	return nil
}

func (s *AttendeeServiceImplData) CanRegisterAtThisTime(ctx context.Context) error {
	group, err := ctxvalues.AuthorizedAsGroup(ctx)
	if err != nil || (group != config.TokenForAdmin && group != config.OptionalTokenForInitialReg) {
		// staff and admin may always register, but regular people have to wait until the registration start time
		current := time.Now()
		target := config.RegistrationStartTime()
		secondsToGo := target.Sub(current).Seconds()
		if secondsToGo > 0 {
			return errors.New("public registration has not opened at this time, please come back later")
		}
	}

	return nil
}

func isDuplicateAttendee(ctx context.Context, nickname string, zip string, email string, expectedCount int64) (bool, error) {
	count, err := database.GetRepository().CountAttendeesByNicknameZipEmail(ctx, nickname, zip, email)
	if err != nil {
		return false, err
	}
	return count != expectedCount, nil
}

func checkNoForbiddenChanges(ctx context.Context, key string, choiceConfig config.ChoiceConfig, originalChoices map[string]bool, newChoices map[string]bool) error {
	if choiceConfig.AdminOnly || choiceConfig.ReadOnly {
		if originalChoices[key] != newChoices[key] {
			group, err := ctxvalues.AuthorizedAsGroup(ctx)
			if err != nil || group != config.TokenForAdmin {
				return errors.New("forbidden change in state of choice key " + key + " - only an admin can do that")
			}
		}
	}
	return nil
}

func checkNoConstraintViolation(key string, choiceConfig config.ChoiceConfig, newChoices map[string]bool) error {
	if choiceConfig.Constraint != "" {
		constraints := strings.Split(choiceConfig.Constraint, ",")
		for _, cn := range constraints {
			constraintK := cn
			if strings.HasPrefix(cn, "!") {
				constraintK = strings.TrimPrefix(cn, "!")
				if newChoices[key] && newChoices[constraintK] {
					return errors.New("cannot pick both " + key + " and " + constraintK + " - constraint violated")
				}
			} else {
				if newChoices[key] && !newChoices[constraintK] {
					return errors.New("when picking " + key + ", must also pick " + constraintK + " - constraint violated")
				}
			}
		}
	}
	return nil
}

func choiceStrToMap(choiceStr string) map[string]bool {
	result := make(map[string]bool)
	if choiceStr != "" {
		choices := strings.Split(choiceStr, ",")
		for _, pickedKey := range choices {
			result[pickedKey] = true
		}
	}
	return result
}
