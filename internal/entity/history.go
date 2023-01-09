package entity

import "gorm.io/gorm"

type History struct {
	gorm.Model
	Entity    string `gorm:"type:varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:att_histories_entity_idx"`    // the name (type) of the entity
	EntityId  uint   `gorm:"NOT NULL;index:att_histories_entity_idx"`                                                                      // the pk of the entity
	RequestId string `gorm:"type:varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`                                             // optional request id that triggered the change
	Identity  string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;NOT NULL;index:att_histories_identity_idx"` // the subject that triggered the change
	Diff      string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
}
