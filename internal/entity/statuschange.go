package entity

import (
	"github.com/eurofurence/reg-attendee-service/internal/api/v1/status"
	"gorm.io/gorm"
)

// configured sizes are for mysql, since version 5 mysql counts characters, not bytes

type StatusChange struct {
	gorm.Model
	AttendeeId uint          `gorm:"NOT NULL;index:att_status_changes_attendee_idx"`
	Status     status.Status `gorm:"type:varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:att_status_changes_status_idx"`
	Comments   string        `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
}
