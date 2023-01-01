package entity

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"gorm.io/gorm"
)

// configured sizes are for mysql, since version 5 mysql counts characters, not bytes

type Attendee struct {
	gorm.Model
	Nickname             string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:nick_idx"`
	FirstName            string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	LastName             string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	Street               string `gorm:"type:varchar(120) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	Zip                  string `gorm:"type:varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	City                 string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	Country              string `gorm:"type:varchar(2) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	State                string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	Email                string `gorm:"type:varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:email_idx"`
	Phone                string `gorm:"type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	Telegram             string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	Partner              string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	Birthday             string `gorm:"type:varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	Gender               string `gorm:"type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL"`
	Pronouns             string `gorm:"type:varchar(40) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	TshirtSize           string `gorm:"type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	SpokenLanguages      string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"` // comma-separated choice field with leading and trailing comma
	RegistrationLanguage string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"` // comma-separated choice field with leading and trailing comma
	Flags                string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"` // comma-separated choice field with leading and trailing comma
	Packages             string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"` // comma-separated choice field with leading and trailing comma
	Options              string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"` // comma-separated choice field with leading and trailing comma
	UserComments         string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci" testdiff:"ignore"`
	Identity             string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	CacheTotalDues       int64  // cache for search functionality only: valid dues balance
	CachePaymentBalance  int64  // cache for search functionality only: valid payments balance
	CacheOpenBalance     int64  // cache for search functionality only: tentative + pending payments balance
	CacheDueDate         string `gorm:"type:varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"` // cache for search functionality only: iso due date
}

type AttendeeQueryResult struct {
	Attendee
	Status        status.Status
	AdminComments string
	AdminFlags    string
}
