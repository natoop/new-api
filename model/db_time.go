package model

import (
	"github.com/QuantumNous/new-api/common"

	"gorm.io/gorm"
)

// GetDBTimestamp returns a UNIX timestamp from database time.
// Falls back to application time on error.
func GetDBTimestamp() int64 {
	return getDBTimestampWith(DB)
}

// getDBTimestampWith 与 GetDBTimestamp 一致，但允许指定连接/事务。
// 在事务内必须传入 tx，避免占用第二个连接（单连接池下会死锁）。
func getDBTimestampWith(db *gorm.DB) int64 {
	if db == nil {
		db = DB
	}
	var ts int64
	var err error
	switch {
	case common.UsingPostgreSQL:
		err = db.Raw("SELECT EXTRACT(EPOCH FROM NOW())::bigint").Scan(&ts).Error
	case common.UsingSQLite:
		err = db.Raw("SELECT strftime('%s','now')").Scan(&ts).Error
	default:
		err = db.Raw("SELECT UNIX_TIMESTAMP()").Scan(&ts).Error
	}
	if err != nil || ts <= 0 {
		return common.GetTimestamp()
	}
	return ts
}
