package mysqldb

import (
	"context"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/entity"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/config"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/database/dbrepo"
	"github.com/jumpy-squirrel/rexis-go-attendee/internal/repository/logging"
)

type MysqlRepository struct {
	db *gorm.DB
}

func Create() dbrepo.Repository {
	return &MysqlRepository{}
}

func (r *MysqlRepository) Open() {
	db, err := gorm.Open("mysql", config.DatabaseMysqlConnectString())
	if err != nil {
		logging.NoCtx().Fatalf("failed to open mysql connection: %v", err)
	}
	r.db = db
}

func (r *MysqlRepository) Close() {
	err := r.db.Close()
	if err != nil {
		logging.NoCtx().Fatalf("failed to close mysql connection: %v", err)
	}
}

func (r *MysqlRepository) Migrate() {
	err := r.db.AutoMigrate(&entity.Attendee{}, &entity.History{}).Error
	if err != nil {
		logging.NoCtx().Fatalf("failed to migrate mysql db: %v", err)
	}
}

func (r *MysqlRepository) AddAttendee(ctx context.Context, a *entity.Attendee) (uint, error) {
	err := r.db.Create(a).Error
	if err != nil {
		logging.Ctx(ctx).Warnf("mysql error during attendee insert: %v", err)
	}
	return a.ID, err
}

func (r *MysqlRepository) UpdateAttendee(ctx context.Context, a *entity.Attendee) error {
	err := r.db.Save(a).Error
	if err != nil {
		logging.Ctx(ctx).Warnf("mysql error during attendee update: %v", err)
	}
	return err
}

func (r *MysqlRepository) GetAttendeeById(ctx context.Context, id uint) (*entity.Attendee, error) {
	var a entity.Attendee
	err := r.db.First(&a, id).Error
	if err != nil {
		logging.Ctx(ctx).Infof("mysql error during attendee select - might be ok: %v", err)
	}
	return &a, err
}

func (r *MysqlRepository) CountAttendeesByNicknameZipEmail(ctx context.Context, nickname string, zip string, email string) (int64, error) {
	var count int64
	err := r.db.Table("attendees").Where(&entity.Attendee{Nickname: nickname, Zip: zip, Email: email}).Count(&count).Error
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (r *MysqlRepository) RecordHistory(ctx context.Context, h *entity.History) error {
	err := r.db.Create(h).Error
	if err != nil {
		logging.Ctx(ctx).Warnf("mysql error during history entry insert: %v", err)
	}
	return err
}
