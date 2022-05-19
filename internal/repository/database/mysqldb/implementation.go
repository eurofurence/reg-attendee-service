package mysqldb

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/entity"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/database/dbrepo"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

type MysqlRepository struct {
	db *gorm.DB
}

func Create() dbrepo.Repository {
	return &MysqlRepository{}
}

func (r *MysqlRepository) Open() {
	gormConfig := gorm.Config{}
	connectString := config.DatabaseMysqlConnectString()

	db, err := gorm.Open(mysql.Open(connectString), &gormConfig)
	if err != nil {
		logging.NoCtx().Fatalf("failed to open mysql connection: %v", err)
	}

	sqlDb, err := db.DB()
	if err != nil {
		logging.NoCtx().Fatalf("failed to configure mysql connection: %v", err)
	}

	// see https://making.pusher.com/production-ready-connection-pooling-in-go/
	sqlDb.SetMaxOpenConns(100)
	sqlDb.SetMaxIdleConns(50)
	sqlDb.SetConnMaxLifetime(time.Minute * 10)

	r.db = db
}

func (r *MysqlRepository) Close() {
	// no more db close in gorm v2
}

func (r *MysqlRepository) Migrate() {
	err := r.db.AutoMigrate(&entity.Attendee{}, &entity.History{}, &entity.AdminInfo{}, &entity.StatusChange{}).Error
	if err != nil {
		logging.NoCtx().Fatalf("failed to migrate mysql db: %v", err)
	}
}

// --- attendee ---

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
	err := r.db.Model(&entity.Attendee{}).Where(&entity.Attendee{Nickname: nickname, Zip: zip, Email: email}).Count(&count).Error
	if err != nil {
		return -1, err
	}
	return count, nil
}

func (r *MysqlRepository) MaxAttendeeId(ctx context.Context) (uint, error) {
	var max uint
	rows, err := r.db.Model(&entity.Attendee{}).Select("ifnull(max(id),0) AS max_id").Rows()
	if err != nil {
		logging.Ctx(ctx).Error("error querying for max attendee id: " + err.Error())
		return 0, err
	}
	for rows.Next() {
		err = rows.Scan(&max)
		if err != nil {
			logging.Ctx(ctx).Error("error reading max attendee id: " + err.Error())
			break
		}
	}
	err2 := rows.Close()
	if err2 != nil {
		logging.Ctx(ctx).Warn("secondary error closing recordset: " + err2.Error())
	}
	return max, err
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
			logging.Ctx(ctx).Infof("mysql error during admin info select - not record not found: %v", err)
			ai.ID = attendeeId
		}
	}
	return &ai, err
}

func (r *MysqlRepository) WriteAdminInfo(ctx context.Context, ai *entity.AdminInfo) error {
	err := r.db.Save(ai).Error
	if err != nil {
		logging.Ctx(ctx).Warnf("mysql error during admin info save: %v", err)
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
				Status:     "new",
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
		logging.Ctx(ctx).Warnf("mysql error during status change select: %v", err)
		return make([]entity.StatusChange, 0), err
	}
	defer func() {
		err := rows.Close()
		logging.Ctx(ctx).Warnf("mysql error during status change result set close: %v", err)
	}()

	result := make([]entity.StatusChange, 0)
	for rows.Next() {
		var sc entity.StatusChange
		err := r.db.ScanRows(rows, &sc)
		if err != nil {
			logging.Ctx(ctx).Warnf("mysql error during status change read: %v", err)
			return make([]entity.StatusChange, 0), err
		}

		result = append(result, sc)
	}

	return result, nil
}

func (r *MysqlRepository) AddStatusChange(ctx context.Context, sc *entity.StatusChange) error {
	err := r.db.Create(sc).Error
	if err != nil {
		logging.Ctx(ctx).Warnf("mysql error during status change insert: %v", err)
	}
	return err
}

// --- history ---

func (r *MysqlRepository) RecordHistory(ctx context.Context, h *entity.History) error {
	err := r.db.Create(h).Error
	if err != nil {
		logging.Ctx(ctx).Warnf("mysql error during history entry insert: %v", err)
	}
	return err
}
