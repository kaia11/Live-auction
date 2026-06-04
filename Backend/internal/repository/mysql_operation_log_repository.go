package repository

import (
	"database/sql"

	"auction-live/backend/internal/model"
)

type MySQLOperationLogRepository struct {
	db *sql.DB
}

func NewMySQLOperationLogRepository(db *sql.DB) *MySQLOperationLogRepository {
	return &MySQLOperationLogRepository{db: db}
}

func (r *MySQLOperationLogRepository) CreateLog(log model.OperationLog) error {
	_, err := r.db.Exec(`
		INSERT INTO operation_logs (module, action, operator_id, room_id, target_type, target_id, content, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, log.Module, log.Action, log.OperatorID, nullableEmptyToNull(log.RoomID), log.TargetType, log.TargetID, nullableEmptyToNull(log.Content), nullableTimeValue(log.CreateTime))
	return err
}
