package entity

import "gorm.io/gorm"

// configured sizes are for mysql, since version 5 mysql counts characters, not bytes

type AdminInfo struct {
	gorm.Model
	Flags                 string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	Permissions           string `gorm:"type:varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
	AdminComments         string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci" testdiff:"ignore"`
	ManualDues            int64
	ManualDuesDescription string `gorm:"type:text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"`
}
