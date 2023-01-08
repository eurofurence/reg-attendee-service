package entity

import "gorm.io/gorm"

// configured sizes are for mysql, since version 5 mysql counts characters, not bytes

type AdditionalInfo struct {
	gorm.Model
	AttendeeId uint   `gorm:"NOT NULL;uniqueIndex:att_add_infos_area_uidx"`
	Area       string `gorm:"type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;uniqueIndex:att_add_infos_area_uidx"`
	JsonValue  string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
}
