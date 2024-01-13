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
	"github.com/eurofurence/reg-attendee-service/internal/web/util/ctxvalues"
	"sort"
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
		aulogging.Logger.Ctx(ctx).Warn().Printf("received new registration duplicate - nick %s zip %s email %s", attendee.Nickname, attendee.Zip, attendee.Email)
		return 0, errors.New("duplicate attendee data - you are already registered")
	}

	// record which user owns this attendee
	attendee.Identity = ctxvalues.Subject(ctx)

	if config.RequireLoginForReg() {
		alreadyHasRegistration, err := userAlreadyHasAnotherRegistration(ctx, attendee.Identity, 0)
		if err != nil {
			return 0, err
		}
		if alreadyHasRegistration {
			aulogging.Logger.Ctx(ctx).Warn().Printf("received second registration for same user - nick %s zip %s email %s", attendee.Nickname, attendee.Zip, attendee.Email)
			return 0, errors.New("duplicate - must use a separate email address and identity account for each person")
		}
	}

	attendee.Flags = s.setAutoFlags(ctx, attendee.Flags)

	id, err := database.GetRepository().AddAttendee(ctx, attendee)
	return id, err
}

func (s *AttendeeServiceImplData) setAutoFlags(ctx context.Context, flags string) string {
	for key, conf := range config.FlagsConfigNoAdmin() {
		if conf.Group != "" {
			if ctxvalues.IsAuthorizedAsGroup(ctx, conf.Group) {
				if !strings.Contains(flags, ","+key+",") {
					flags += key + ","
				}
			}
		}
	}
	return flags
}

func (s *AttendeeServiceImplData) GetAttendee(ctx context.Context, id uint) (*entity.Attendee, error) {
	attendee, err := database.GetRepository().GetAttendeeById(ctx, id)
	return attendee, err
}

func (s *AttendeeServiceImplData) UpdateAttendee(ctx context.Context, attendee *entity.Attendee, suppressMinorUpdateEmails bool) error {
	alreadyExists, err := isDuplicateAttendee(ctx, attendee.Nickname, attendee.Zip, attendee.Email, 1)
	if err != nil {
		return err
	}
	if alreadyExists {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received update with registration duplicate - nick %s zip %s email %s", attendee.Nickname, attendee.Zip, attendee.Email)
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

	subject := ctxvalues.Subject(ctx)
	// changing packages may change the due amount
	err = s.UpdateDuesAndDoStatusChangeIfNeeded(ctx, attendee, currentStatus, currentStatus, fmt.Sprintf("attendee update by %s", subject), "", suppressMinorUpdateEmails)
	if err != nil {
		return err
	}

	return nil
}

func (s *AttendeeServiceImplData) GetAttendeeMaxId(ctx context.Context) (uint, error) {
	max, err := database.GetRepository().MaxAttendeeId(ctx)
	return max, err
}

func (s *AttendeeServiceImplData) CanChangeEmailTo(ctx context.Context, originalEmail string, newEmail string) error {
	if !config.RequireLoginForReg() {
		// cannot validate here, need separate validation step
		return nil
	}

	if originalEmail == newEmail {
		// allow even normal users to keep an email once set by an admin
		return nil
	}

	if ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup()) || ctxvalues.HasApiToken(ctx) {
		// allow admins or api token to set anything
		return nil
	}

	if !ctxvalues.EmailVerified(ctx) {
		return errors.New("you must verify your email address with the identity provider first")
	}

	if ctxvalues.Email(ctx) == newEmail {
		// anyone can set their own email address, as validated by IDP - we already know not empty
		return nil
	}

	return errors.New("you can only use the email address you're logged in with")
}

func (s *AttendeeServiceImplData) CanChangeChoiceTo(ctx context.Context, what string, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig) error {
	return s.CanChangeChoiceToCurrentStatus(ctx, what, originalChoiceStr, newChoiceStr, configuration, "irrelevant")
}

func (s *AttendeeServiceImplData) CanChangeChoiceToCurrentStatus(ctx context.Context, what string, originalChoiceStr string, newChoiceStr string, configuration map[string]config.ChoiceConfig, currentStatus status.Status) error {
	originalChoices := choiceStrToMap(originalChoiceStr, configuration)
	newChoices := choiceStrToMap(newChoiceStr, configuration)
	oneIsMandatory := false
	satisfiesOneIsMandatory := false
	mandatoryList := make([]string, 0)
	for k, v := range configuration {
		if err := checkNoForbiddenChanges(ctx, what, k, v, originalChoices, newChoices); err != nil {
			return err
		}
		if err := checkNoConstraintViolation(k, v, newChoices); err != nil {
			return err
		}
		if currentStatus != "irrelevant" {
			if err := checkNoForbiddenChangesAfterPayment(ctx, what, k, v, configuration, originalChoices, newChoices, currentStatus); err != nil {
				return err
			}
		}
		if v.Mandatory {
			oneIsMandatory = true
			mandatoryList = append(mandatoryList, k)
			if newChoices[k] {
				satisfiesOneIsMandatory = true
			}
		}
	}

	if oneIsMandatory && !satisfiesOneIsMandatory {
		sort.Strings(mandatoryList)
		return fmt.Errorf("you must pick at least one of the mandatory options (%s)", strings.Join(mandatoryList, ","))
	}

	return nil
}

func (s *AttendeeServiceImplData) CanRegisterAtThisTime(ctx context.Context) error {
	// staff early reg? (also for admins)
	earlyRole := config.OidcEarlyRegGroup()
	if earlyRole != "" && (ctxvalues.IsAuthorizedAsGroup(ctx, earlyRole) || ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup())) {
		current := time.Now()
		target := config.EarlyRegistrationStartTime()
		secondsToGo := target.Sub(current).Seconds()
		if secondsToGo > 0 {
			return errors.New("staff registration has not opened at this time, please come back later")
		}
		return nil
	}

	// regular people have to wait until the registration start time
	current := time.Now()
	target := config.RegistrationStartTime()
	secondsToGo := target.Sub(current).Seconds()
	if secondsToGo > 0 {
		return errors.New("public registration has not opened at this time, please come back later")
	}
	return nil
}

func isDuplicateAttendee(ctx context.Context, nickname string, zip string, email string, expectedCountMax int64) (bool, error) {
	count, err := database.GetRepository().CountAttendeesByNicknameZipEmail(ctx, nickname, zip, email)
	if err != nil {
		return false, err
	}
	return count > expectedCountMax, nil
}

func userAlreadyHasAnotherRegistration(ctx context.Context, identity string, expectedCount int64) (bool, error) {
	if identity == "" {
		return false, nil
	}

	count, err := database.GetRepository().CountAttendeesByIdentity(ctx, identity)
	if err != nil {
		return false, err
	}
	return count != expectedCount, nil
}

func checkNoForbiddenChanges(ctx context.Context, what string, key string, choiceConfig config.ChoiceConfig, originalChoices map[string]bool, newChoices map[string]bool) error {
	if originalChoices[key] != newChoices[key] {
		// tolerate removing a read-only choice that has a constraint that forbids it anyway
		if choiceConfig.ReadOnly {
			if originalChoices[key] && !newChoices[key] {
				if canAllowRemovalDueToConstraint(ctx, what, key, choiceConfig, originalChoices, newChoices) {
					return nil
				}
			}
		}
		if choiceConfig.AdminOnly || choiceConfig.ReadOnly {
			if !ctxvalues.HasApiToken(ctx) && !ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup()) {
				return fmt.Errorf("forbidden select or deselect of %s %s - only an admin can do that", what, key)
			}
		}
	}
	return nil
}

func canAllowRemovalDueToConstraint(ctx context.Context, what string, key string, choiceConfig config.ChoiceConfig, originalChoices map[string]bool, newChoices map[string]bool) bool {
	if choiceConfig.Constraint != "" {
		constraints := strings.Split(choiceConfig.Constraint, ",")
		for _, cn := range constraints {
			constraintK := cn
			if strings.HasPrefix(cn, "!") {
				constraintK = strings.TrimPrefix(cn, "!")
				if newChoices[constraintK] {
					aulogging.Logger.Ctx(ctx).Info().Printf("can allow removal of read only %s %s - it would violate a constraint for %s anyway", what, key, constraintK)
					return true
				}
			}
		}
	}
	return false
}

func checkNoForbiddenChangesAfterPayment(ctx context.Context, what string, key string, choiceConfig config.ChoiceConfig, configuration map[string]config.ChoiceConfig, originalChoices map[string]bool, newChoices map[string]bool, currentStatus status.Status) error {
	if ctxvalues.HasApiToken(ctx) || ctxvalues.IsAuthorizedAsGroup(ctx, config.OidcAdminGroup()) {
		return nil
	}

	if currentStatus == status.PartiallyPaid || currentStatus == status.Paid || currentStatus == status.CheckedIn {
		if originalChoices[key] && !newChoices[key] && choiceConfig.Price > 0 {
			oldDues := calcTotalDuesHelper(configuration, originalChoices)
			newDues := calcTotalDuesHelper(configuration, newChoices)

			if newDues < oldDues {
				return fmt.Errorf("deselect of %s %s after payment leads to dues reduction - only an admin can do that at this time", what, key)
			}
		}
	}

	return nil
}

func calcTotalDuesHelper(configuration map[string]config.ChoiceConfig, choices map[string]bool) (dues int64) {
	for k, selected := range choices {
		choiceConfig, ok := configuration[k]
		if ok && selected {
			dues += choiceConfig.Price
		}
	}
	return dues
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

func choiceStrToMap(choiceStr string, configuration map[string]config.ChoiceConfig) map[string]bool {
	result := make(map[string]bool)
	// ensure all available keys present
	for k, _ := range configuration {
		result[k] = false
	}
	if choiceStr != "" {
		choices := strings.Split(choiceStr, ",")
		for _, pickedKey := range choices {
			if pickedKey != "" {
				result[pickedKey] = true
			}
		}
	}
	return result
}

func commaSeparatedStrToMap(choiceStr string, allowedValues []string) map[string]bool {
	result := make(map[string]bool)
	// ensure all available values present
	for _, k := range allowedValues {
		result[k] = false
	}
	if choiceStr != "" {
		choices := strings.Split(choiceStr, ",")
		for _, pickedKey := range choices {
			if pickedKey != "" {
				result[pickedKey] = true
			}
		}
	}
	return result
}
