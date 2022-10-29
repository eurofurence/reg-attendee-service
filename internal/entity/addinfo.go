package entity

import "gorm.io/gorm"

// configured sizes are for mysql, since version 5 mysql counts characters, not bytes

type AdditionalInfo struct {
	gorm.Model
	AttendeeId uint   `gorm:"NOT NULL;unique_index:attendee_area_idx"`
	Area       string `gorm:"type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;unique_index:attendee_area_idx"`
	JsonValue  string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
}
