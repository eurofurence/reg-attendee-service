package entity

import "github.com/jinzhu/gorm"

type Attendee struct {
	gorm.Model
	Nickname     string
	FirstName    string
	LastName     string
	Street       string
	Zip          string
	City         string
	Country      string
	CountryBadge string
	State        string
	Email        string
	Phone        string
	Telegram     string
	Birthday     string
	Gender       string
	TshirtSize   string
	Flags        string
	Packages     string
	Options      string
	UserComments string
}
