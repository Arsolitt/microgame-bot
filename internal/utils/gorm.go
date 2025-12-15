package utils

import (
	"gorm.io/gorm"
)

func IsInGormTransaction(db *gorm.DB) bool {
	committer, ok := db.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}
