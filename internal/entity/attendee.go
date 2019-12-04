package entity

import "github.com/jinzhu/gorm"

// configured sizes are for mysql, since version 5 mysql counts characters, not bytes

type Attendee struct {
	gorm.Model
	Nickname     string `gorm:"type:varchar(80);NOT NULL;index:nick_idx"`
	FirstName    string `gorm:"type:varchar(80);NOT NULL"`
	LastName     string `gorm:"type:varchar(80);NOT NULL"`
	Street       string `gorm:"type:varchar(120);NOT NULL"`
	Zip          string `gorm:"type:varchar(20);NOT NULL"`
	City         string `gorm:"type:varchar(80);NOT NULL"`
	Country      string `gorm:"type:varchar(2);NOT NULL"`
	CountryBadge string `gorm:"type:varchar(2);NOT NULL"`
	State        string `gorm:"type:varchar(80)"`
	Email        string `gorm:"type:varchar(200);NOT NULL;index:email_idx"`
	Phone        string `gorm:"type:varchar(32);NOT NULL"`
	Telegram     string `gorm:"type:varchar(80)"`
	Birthday     string `gorm:"type:varchar(10);NOT NULL"`
	Gender       string `gorm:"type:varchar(32);NOT NULL"`
	TshirtSize   string `gorm:"type:varchar(32)"`
	Flags        string `gorm:"type:varchar(255)"`
	Packages     string `gorm:"type:varchar(255)"`
	Options      string `gorm:"type:varchar(255)"`
	UserComments string `gorm:"type:text" testdiff:"ignore"`
}
