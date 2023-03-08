package attendeesrv

import (
	"context"
	"errors"
	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database"
	"regexp"
	"strings"
)

var DuplicateBanError = errors.New("duplicate ban rule")

func (s *AttendeeServiceImplData) NewBan(ctx context.Context) *entity.Ban {
	return &entity.Ban{}
}

func (s *AttendeeServiceImplData) CreateBan(ctx context.Context, ban *entity.Ban) (uint, error) {
	if ban.ID != 0 {
		aulogging.Logger.Ctx(ctx).Error().Print("cannot create ban rule with assigned id - this is a program error")
		return ban.ID, errors.New("cannot create ban rule with assigned id - this is a program error")
	}

	alreadyExists, err := isDuplicateBan(ctx, ban)
	if err != nil {
		return 0, err
	}
	if alreadyExists {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received new ban rule duplicate - name_pattern %s nickname_pattern %s email_pattern %s", ban.NamePattern, ban.NicknamePattern, ban.EmailPattern)
		return 0, DuplicateBanError
	}

	id, err := database.GetRepository().AddBan(ctx, ban)
	return id, err
}

func (s *AttendeeServiceImplData) UpdateBan(ctx context.Context, ban *entity.Ban) error {
	if ban.ID == 0 {
		aulogging.Logger.Ctx(ctx).Error().Print("cannot update ban rule without assigned id - this is a program error")
		return errors.New("cannot update ban rule without assigned id - this is a program error")
	}

	alreadyExists, err := isDuplicateBan(ctx, ban)
	if err != nil {
		return err
	}
	if alreadyExists {
		aulogging.Logger.Ctx(ctx).Warn().Printf("received update that would lead to ban rule duplicate - name_pattern %s nickname_pattern %s email_pattern %s", ban.NamePattern, ban.NicknamePattern, ban.EmailPattern)
		return DuplicateBanError
	}

	err = database.GetRepository().UpdateBan(ctx, ban)
	return err
}

func (s *AttendeeServiceImplData) DeleteBan(ctx context.Context, ban *entity.Ban) error {
	if ban.ID == 0 {
		aulogging.Logger.Ctx(ctx).Error().Print("cannot delete ban rule without assigned id - this is a program error")
		return errors.New("cannot delete ban rule without assigned id - this is a program error")
	}

	err := database.GetRepository().DeleteBan(ctx, ban)
	return err
}

func (s *AttendeeServiceImplData) GetBan(ctx context.Context, id uint) (*entity.Ban, error) {
	return database.GetRepository().GetBanById(ctx, id)
}

func (s *AttendeeServiceImplData) GetAllBans(ctx context.Context) ([]*entity.Ban, error) {
	return database.GetRepository().GetAllBans(ctx)
}

func isDuplicateBan(ctx context.Context, ban *entity.Ban) (bool, error) {
	currentBans, err := database.GetRepository().GetAllBans(ctx)
	if err != nil {
		return false, err
	}

	for _, b := range currentBans {
		if b.ID != ban.ID {
			if b.NamePattern == ban.NamePattern &&
				b.NicknamePattern == ban.NicknamePattern &&
				b.EmailPattern == ban.EmailPattern {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *AttendeeServiceImplData) matchesBanAndNoSkip(ctx context.Context, attendee *entity.Attendee) error {
	adminInfo, err := s.GetAdminInfo(ctx, attendee.ID)
	if err != nil {
		return err
	}

	if strings.Contains(adminInfo.Flags, ",skip_ban_check,") {
		return nil
	}

	bans, err := s.GetAllBans(ctx)
	if err != nil {
		return err
	}

	for _, ban := range bans {
		if matchesLowercased(ctx, "nickname", ban.ID, ban.NicknamePattern, attendee.Nickname) ||
			matchesLowercased(ctx, "name", ban.ID, ban.NamePattern, attendee.FirstName+" "+attendee.LastName) ||
			matchesLowercased(ctx, "email", ban.ID, ban.EmailPattern, attendee.Email) {
			return BanCandidateError
		}
	}

	return nil
}

func matchesLowercased(ctx context.Context, field string, id uint, pattern string, value string) bool {
	match, err := regexp.MatchString(pattern, strings.ToLower(value))
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().Printf("invalid %s pattern '%s' in ban %d - ignored due to compile error: %s", field, pattern, id, err.Error())
		return false
	}
	return match
}
