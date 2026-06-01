package repository

import (
	"database/sql"
	"fmt"
	"time"

	"auction-live/backend/internal/model"
)

type MySQLResultRepository struct {
	db *sql.DB
}

func NewMySQLResultRepository(db *sql.DB) *MySQLResultRepository {
	return &MySQLResultRepository{db: db}
}

func (r *MySQLResultRepository) CreateResult(result model.AuctionResult) error {
	resultID := fmt.Sprintf("result-%s", result.SessionID)
	_, err := r.db.Exec(`
		INSERT INTO auction_results (id, session_id, item_id, result_status, winner_user_id, final_price, participant_count, ended_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			result_status = VALUES(result_status),
			winner_user_id = VALUES(winner_user_id),
			final_price = VALUES(final_price),
			participant_count = VALUES(participant_count),
			ended_at = VALUES(ended_at)
	`, resultID, result.SessionID, result.ItemID, result.ResultStatus, result.WinnerUserID, result.FinalPrice, result.ParticipantCount, time.Now().Format("2006-01-02 15:04:05"))
	return err
}
