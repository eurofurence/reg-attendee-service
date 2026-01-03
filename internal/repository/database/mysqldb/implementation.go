package mysqldb

import (
	"context"
	"errors"
	"fmt"
	"time"

	aulogging "github.com/StephanHCB/go-autumn-logging"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/attendee"
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/dbrepo"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type MysqlRepository struct {
	db  *gorm.DB
	Now func() time.Time
}

func Create() dbrepo.Repository {
	return &MysqlRepository{
		Now: time.Now,
	}
}

func (r *MysqlRepository) Open() error {
	gormConfig := gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "att_",
		},
		Logger: logger.Default.LogMode(logger.Silent),
	}
	connectString := config.DatabaseMysqlConnectString()

	db, err := gorm.Open(mysql.Open(connectString), &gormConfig)
	if err != nil {
		aulogging.Logger.NoCtx().Error().WithErr(err).Printf("failed to open mysql connection: %s", err.Error())
		return err
	}

	sqlDb, err := db.DB()
	if err != nil {
		aulogging.Logger.NoCtx().Error().WithErr(err).Printf("failed to configure mysql connection: %s", err.Error())
		return err
	}

	// see https://making.pusher.com/production-ready-connection-pooling-in-go/
	sqlDb.SetMaxOpenConns(100)
	sqlDb.SetMaxIdleConns(50)
	sqlDb.SetConnMaxLifetime(time.Minute * 10)

	r.db = db
	return nil
}

func (r *MysqlRepository) Close() {
	// no more db close in gorm v2
}

func (r *MysqlRepository) Migrate() error {
	err := r.db.AutoMigrate(
		&entity.AdditionalInfo{},
		&entity.AdminInfo{},
		&entity.Attendee{},
		&entity.Ban{},
		&entity.History{},
		&entity.StatusChange{},
		&entity.Count{},
	)
	if err != nil {
		aulogging.Logger.NoCtx().Error().WithErr(err).Printf("failed to migrate mysql db: %s", err.Error())
		return err
	}
	return nil
}

// --- attendee ---

func (r *MysqlRepository) AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error) {
	err := r.db.Create(a).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee insert: %s", err.Error())
	}
	return a.ID, err
}

func (r *MysqlRepository) UpdateAttendee(ctx context.Context, a *entity.Attendee) error {
	// allow updating deleted (because the admin ui allows it)
	err := r.db.Unscoped().Save(a).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee update: %s", err.Error())
	}
	return err
}

func (r *MysqlRepository) GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error) {
	var a entity.Attendee
	// allow reading deleted so history and undelete work
	err := r.db.Unscoped().First(&a, id).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Info().WithErr(err).Printf("mysql error during attendee select - might be ok: %s", err.Error())
	}
	return &a, err
}

func (r *MysqlRepository) SoftDeleteAttendeeById(ctx context.Context, id uint) error {
	var a entity.Attendee
	err := r.db.First(&a, id).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee soft delete - attendee not found: %s", err.Error())
		return err
	}
	err = r.db.Delete(&a).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee soft delete - deletion failed: %s", err.Error())
		return err
	}
	return nil
}

func (r *MysqlRepository) UndeleteAttendeeById(ctx context.Context, id uint) error {
	var a entity.Attendee
	err := r.db.Unscoped().First(&a, id).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee undelete - attendee not found: %s", err.Error())
		return err
	}
	err = r.db.Unscoped().Model(&a).Where("id", id).Update("deleted_at", nil).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee undelete: %s", err.Error())
		return err
	}
	return nil
}

func (r *MysqlRepository) CountAttendeesByNicknameZipEmail(ctx context.Context, nickname string, zip string, email string) (int64, error) {
	var count int64
	// count deleted because the unique index in the db will
	err := r.db.Unscoped().Model(&entity.Attendee{}).Where(&entity.Attendee{Nickname: nickname, Zip: zip, Email: email}).Count(&count).Error
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (r *MysqlRepository) CountAttendeesByIdentity(ctx context.Context, identity string) (int64, error) {
	var count int64
	// count deleted because the unique index in the db will
	err := r.db.Unscoped().Model(&entity.Attendee{}).Where(&entity.Attendee{Identity: identity}).Count(&count).Error
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (r *MysqlRepository) MaxAttendeeId(ctx context.Context) (uint, error) {
	var max uint
	// count deleted
	rows, err := r.db.Unscoped().Model(&entity.Attendee{}).Select("ifnull(max(id),0) AS max_id").Rows()
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error querying for max attendee id: %s", err.Error())
		return 0, err
	}
	for rows.Next() {
		err = rows.Scan(&max)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading max attendee id: %s", err.Error())
			break
		}
	}
	err2 := rows.Close()
	if err2 != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err2).Printf("secondary error closing recordset: %s", err2.Error())
	}
	return max, err
}

// --- attendee search ---

func (r *MysqlRepository) FindAttendees(ctx context.Context, criteria *attendee.AttendeeSearchCriteria) ([]*entity.AttendeeQueryResult, error) {
	params := make(map[string]interface{})
	query := r.constructAttendeeSearchQuery(ctx, criteria, params)

	result := make([]*entity.AttendeeQueryResult, 0)

	// Raw finds deleted attendees
	rows, err := r.db.Raw(query, params).Rows()
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error finding attendees: %s", err.Error())
		return result, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err2).Printf("secondary error closing recordset during find: %s", err2.Error())
		}
	}()

	for rows.Next() {
		attendeeBuffer := entity.AttendeeQueryResult{}
		err = r.db.ScanRows(rows, &attendeeBuffer)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading attendeeBuffer during find: %s", err.Error())
			return result, err
		}
		copiedAttendee := attendeeBuffer
		result = append(result, &copiedAttendee)
	}

	return result, nil
}

// --- admin info ---

func (r *MysqlRepository) GetAdminInfoByAttendeeId(ctx context.Context, attendeeId uint) (*entity.AdminInfo, error) {
	var ai entity.AdminInfo
	err := r.db.First(&ai, attendeeId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ai.ID = attendeeId
			err = nil // acceptable situation - we only write admin info on change
		} else {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during admin info select - not record not found: %s", err.Error())
			ai.ID = attendeeId
		}
	}
	return &ai, err
}

func (r *MysqlRepository) WriteAdminInfo(ctx context.Context, ai *entity.AdminInfo) error {
	err := r.db.Save(ai).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during admin info save: %s", err.Error())
	}
	return err
}

// --- status changes ---

func (r *MysqlRepository) GetLatestStatusChangeByAttendeeId(ctx context.Context, attendeeId uint) (*entity.StatusChange, error) {
	var sc entity.StatusChange
	err := r.db.Model(&entity.StatusChange{}).Where(&entity.StatusChange{AttendeeId: attendeeId}).Last(&sc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sc = entity.StatusChange{
				AttendeeId: attendeeId,
				Status:     status.New,
				Comments:   "",
			}
			err = nil
		}
	}
	return &sc, err
}

func (r *MysqlRepository) GetStatusChangesByAttendeeId(ctx context.Context, attendeeId uint) ([]entity.StatusChange, error) {
	rows, err := r.db.Model(&entity.StatusChange{}).Where(&entity.StatusChange{AttendeeId: attendeeId}).Rows()
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during status change select: %s", err.Error())
		return make([]entity.StatusChange, 0), err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during status change result set close: %s", err.Error())
		}
	}()

	result := make([]entity.StatusChange, 0)
	for rows.Next() {
		var sc entity.StatusChange
		err := r.db.ScanRows(rows, &sc)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during status change read: %s", err.Error())
			return make([]entity.StatusChange, 0), err
		}

		result = append(result, sc)
	}

	return result, nil
}

func (r *MysqlRepository) AddStatusChange(ctx context.Context, sc *entity.StatusChange) error {
	err := r.db.Create(sc).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during status change insert: %s", err.Error())
	}
	return err
}

func (r *MysqlRepository) FindByIdentity(ctx context.Context, identity string) ([]*entity.Attendee, error) {
	result := make([]*entity.Attendee, 0)
	rows, err := r.db.Model(&entity.Attendee{}).Where(&entity.Attendee{Identity: identity}).Rows()
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during identity select: %s", err.Error())
		return result, err
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee by identity result set close: %s", err.Error())
		}
	}()

	for rows.Next() {
		var a entity.Attendee
		err := r.db.ScanRows(rows, &a)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during attendee by identity read: %s", err.Error())
			return make([]*entity.Attendee, 0), err
		}

		result = append(result, &a)
	}

	return result, nil
}

// --- bans ---

func (r *MysqlRepository) GetAllBans(ctx context.Context) ([]*entity.Ban, error) {
	result := make([]*entity.Ban, 0)
	banBuffer := entity.Ban{}

	rows, err := r.db.Model(&entity.Ban{}).Order("id").Rows()
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading bans: %s", err.Error())
		return result, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err2).Printf("secondary error closing recordset during find: %s", err2.Error())
		}
	}()

	for rows.Next() {
		err = r.db.ScanRows(rows, &banBuffer)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading ban during find: %s", err.Error())
			return result, err
		}
		copiedBan := banBuffer
		result = append(result, &copiedBan)
	}

	return result, nil
}

func (r *MysqlRepository) GetBanById(ctx context.Context, id uint) (*entity.Ban, error) {
	var b entity.Ban
	err := r.db.First(&b, id).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Info().WithErr(err).Printf("mysql error during ban select - might be ok: %s", err.Error())
	}
	return &b, err
}

func (r *MysqlRepository) AddBan(ctx context.Context, b *entity.Ban) (uint, error) {
	err := r.db.Create(b).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during ban insert: %s", err.Error())
	}
	return b.ID, err
}

func (r *MysqlRepository) UpdateBan(ctx context.Context, b *entity.Ban) error {
	err := r.db.Save(b).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during ban update: %s", err.Error())
	}
	return err
}

func (r *MysqlRepository) DeleteBan(ctx context.Context, b *entity.Ban) error {
	err := r.db.Delete(b).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during ban delete: %s", err.Error())
	}
	return err
}

// --- additional info ---

func (r *MysqlRepository) GetAllAdditionalInfoForArea(ctx context.Context, area string) ([]*entity.AdditionalInfo, error) {
	result := make([]*entity.AdditionalInfo, 0)
	addInfoBuffer := entity.AdditionalInfo{}
	queryBuffer := entity.AdditionalInfo{Area: area}

	rows, err := r.db.Model(&entity.AdditionalInfo{}).Where(&queryBuffer).Order("attendee_id").Rows()
	if err != nil {
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading additional infos for area %s: %s", area, err.Error())
		return result, err
	}
	defer func() {
		err2 := rows.Close()
		if err2 != nil {
			aulogging.Logger.Ctx(ctx).Warn().WithErr(err2).Printf("secondary error closing recordset during additional info read: %s", err2.Error())
		}
	}()

	for rows.Next() {
		err = r.db.ScanRows(rows, &addInfoBuffer)
		if err != nil {
			aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading additional info during find for area %s: %s", area, err.Error())
			return result, err
		}
		copiedAddInfo := addInfoBuffer
		result = append(result, &copiedAddInfo)
	}

	return result, nil
}

func (r *MysqlRepository) GetAdditionalInfoFor(ctx context.Context, attendeeId uint, area string) (*entity.AdditionalInfo, error) {
	var ai entity.AdditionalInfo
	err := r.db.Model(&entity.AdditionalInfo{}).Where(&entity.AdditionalInfo{AttendeeId: attendeeId, Area: area}).Last(&ai).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// return a new entry suitable for saving
			ai = entity.AdditionalInfo{
				AttendeeId: attendeeId,
				Area:       area,
			}
			err = nil
		}
	}
	return &ai, err
}

func (r *MysqlRepository) WriteAdditionalInfo(ctx context.Context, ad *entity.AdditionalInfo) error {
	err := r.db.Save(ad).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during additional info insert or update: %s", err.Error())
	}
	return err
}

// --- count ---

func (r *MysqlRepository) CreateCount(initial *entity.Count) (*entity.Count, error) {
	if initial == nil || initial.Area == "" || initial.Name == "" {
		aulogging.Logger.NoCtx().Error().Print("error setting up counts - received nil or unset area or name")
		return &entity.Count{}, errors.New("error setting up counts - received nil or unset area or name")
	}
	err := r.db.Model(&entity.Count{}).Where(&entity.Count{Area: initial.Area, Name: initial.Name}).FirstOrCreate(initial).Error
	if err != nil {
		aulogging.Logger.NoCtx().Error().WithErr(err).Printf("error setting up counts for area %s name %s: %s", initial.Area, initial.Name, err.Error())
		return initial, err
	}
	return initial, nil
}

func (r *MysqlRepository) AddCount(ctx context.Context, delta *entity.Count) (*entity.Count, error) {
	if delta == nil || delta.Area == "" || delta.Name == "" {
		aulogging.Logger.Ctx(ctx).Error().Print("error changing counts - received nil or unset area or name")
		return &entity.Count{}, errors.New("error changing counts - received nil or unset area or name")
	}
	query := `UPDATE att_counts 
              SET pending = pending + @deltaPending, attending = attending + @deltaAttending 
              WHERE area = @area AND name = @name`
	params := map[string]interface{}{
		"area":           delta.Area,
		"name":           delta.Name,
		"deltaPending":   delta.Pending,
		"deltaAttending": delta.Attending,
	}
	result := r.db.Exec(query, params)
	current, err := r.GetCount(ctx, delta.Area, delta.Name)
	if result.Error != nil {
		aulogging.Logger.Ctx(ctx).Error().Printf("error adding counts area %s name %s: %s", delta.Area, delta.Name, result.Error.Error())
		return current, result.Error
	}
	if result.RowsAffected != 1 {
		aulogging.Logger.Ctx(ctx).Error().Printf("error adding counts area %s name %s: affected %d rows instead of 1", delta.Area, delta.Name, result.RowsAffected)
		return current, fmt.Errorf("error adding counts area %s name %s: failed to affect 1 row", delta.Area, delta.Name)
	}
	return current, err
}

func (r *MysqlRepository) ResetCount(ctx context.Context, overwrite *entity.Count) error {
	if overwrite == nil || overwrite.Area == "" || overwrite.Name == "" {
		aulogging.Logger.Ctx(ctx).Error().Print("error resetting counts - received nil or unset area or name")
		return errors.New("error resetting counts - received nil or unset area or name")
	}
	query := `UPDATE att_counts 
              SET pending = @pending, attending = @attending 
              WHERE area = @area AND name = @name`
	params := map[string]interface{}{
		"area":      overwrite.Area,
		"name":      overwrite.Name,
		"pending":   overwrite.Pending,
		"attending": overwrite.Attending,
	}
	result := r.db.Exec(query, params)
	if result.Error != nil {
		aulogging.Logger.Ctx(ctx).Error().Printf("error resetting counts area %s name %s: %s", overwrite.Area, overwrite.Name, result.Error.Error())
		return result.Error
	}
	if result.RowsAffected != 1 {
		aulogging.Logger.Ctx(ctx).Error().Printf("error adding counts area %s name %s: affected %d rows instead of 1", overwrite.Area, overwrite.Name, result.RowsAffected)
		return fmt.Errorf("error resetting counts area %s name %s: failed to affect 1 row", overwrite.Area, overwrite.Name)
	}
	return nil
}

func (r *MysqlRepository) GetCount(ctx context.Context, area string, name string) (*entity.Count, error) {
	c, err := r.getCountAllowMissing(ctx, area, name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading counts for area %s name %s - should have been set up during database migration: %s", area, name, err.Error())
			return c, err
		}
		aulogging.Logger.Ctx(ctx).Error().WithErr(err).Printf("error reading counts for area %s name %s - other: %s", area, name, err.Error())
		return c, err
	}
	return c, nil
}

func (r *MysqlRepository) getCountAllowMissing(ctx context.Context, area string, name string) (*entity.Count, error) {
	var c entity.Count
	if area == "" || name == "" {
		aulogging.Logger.Ctx(ctx).Error().Print("error reading counts - received unset area or name")
		return &c, errors.New("error reading counts - received unset area or name")
	}
	err := r.db.Model(&entity.Count{}).Where(&entity.Count{Area: area, Name: name}).First(&c).Error
	return &c, err
}

// --- history ---

func (r *MysqlRepository) RecordHistory(ctx context.Context, h *entity.History) error {
	err := r.db.Create(h).Error
	if err != nil {
		aulogging.Logger.Ctx(ctx).Warn().WithErr(err).Printf("mysql error during history entry insert: %s", err.Error())
	}
	return err
}
