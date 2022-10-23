package entity

import "gorm.io/gorm"

type History struct {
	gorm.Model
	Entity    string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:entity_idx"` // the name of the entity
	EntityId  uint   `gorm:"index:entity_idx"`                                                                            // the pk of the entity
	RequestId string `gorm:"type:varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`                            // optional request id that triggered the change
	UserId    string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:user_idx"`
	Diff      string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
}
