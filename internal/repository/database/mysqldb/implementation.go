package mysqldb

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"rexis/rexis-go-attendee/internal/entity"
	"rexis/rexis-go-attendee/internal/repository/config"
)

type MysqlRepository struct {
	db *gorm.DB
}

func (r *MysqlRepository) Open() {
	db, err := gorm.Open("mysql", config.DatabaseMysqlConnectString())
	if err != nil {
		log.Fatalf("failed to open mysql connection: %v", err)
	}
	r.db = db
}

func (r *MysqlRepository) Close() {
	err := r.db.Close()
	if err != nil {
		log.Fatalf("failed to close mysql connection: %v", err)
	}
}

func (r *MysqlRepository) AddAttendee(a *entity.Attendee) (uint, error) {
	err := r.db.Create(a).Error
	return a.ID, err
}

func (r *MysqlRepository) UpdateAttendee(a *entity.Attendee) error {
	err := r.db.Save(a).Error
	return err
}

func (r *MysqlRepository) GetAttendeeById(id uint) (*entity.Attendee, error) {
	var a entity.Attendee
	err := r.db.First(&a, id).Error
	return &a, err
}
