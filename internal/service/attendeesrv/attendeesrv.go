package attendeesrv

import (
	"context"
	"errors"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database"
	"github.com/jumpy-squirrel/rexis-go-attendee/web/filter/ctxvalues"
	"strings"
)

type AttendeeServiceImplData struct {
}

func (s *AttendeeServiceImplData) NewAttendee(ctx context.Context) *entity.Attendee {
	return &entity.Attendee{}
}

func (s *AttendeeServiceImplData) RegisterNewAttendee(ctx context.Context, attendee *entity.Attendee) (uint, error) {
	id, err := database.GetRepository().AddAttendee(ctx, attendee)
	return id, err
}

func (s *AttendeeServiceImplData) GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error) {
	attendee, err := database.GetRepository().GetAttendeeById(ctx, id)
	return attendee, err
}

func (s *AttendeeServiceImplData) UpdateAttendee(ctx context.Context, attendee *entity.Attendee) error {
	err := database.GetRepository().UpdateAttendee(ctx, attendee)
	return err
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
	if (choiceStr != "") {
		choices := strings.Split(choiceStr, ",")
		for _, pickedKey := range choices {
			result[pickedKey] = true
		}
	}
	return result
}
