package entity

import "gorm.io/gorm"

// configured sizes are for mysql, since version 5 mysql counts characters, not bytes

type StatusChange struct {
	gorm.Model
	AttendeeId uint   `gorm:"NOT NULL;index:attendee_id_idx"`
	Status     string `gorm:"type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:status_idx"`
	Comments   string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
}
