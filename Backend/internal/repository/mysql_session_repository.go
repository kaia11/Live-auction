package repository

import (
	"database/sql"

	"auction-live/backend/internal/model"
)

type MySQLSessionRepository struct {
	db *sql.DB
}

func NewMySQLSessionRepository(db *sql.DB) *MySQLSessionRepository {
	return &MySQLSessionRepository{db: db}
}

func (r *MySQLSessionRepository) GetCurrentSession(roomID string) (*model.AuctionSession, error) {
	row := r.db.QueryRow(`
		SELECT s.id, s.room_id, s.item_id, s.status, s.current_price, s.leader_user_id, s.start_time, s.end_time,
		       s.participant_count, s.viewer_count, s.increment_step, s.extension_seconds, s.extension_trigger_seconds,
		       s.ceiling_price, s.supports_auto_proxy
		FROM auction_sessions s
		INNER JOIN live_rooms r ON r.current_session_id = s.id
		WHERE r.id = ?
	`, roomID)

	return scanSession(row)
}

func (r *MySQLSessionRepository) ListRoomSessions(roomID string) ([]model.AuctionSession, error) {
	rows, err := r.db.Query(`
		SELECT id, room_id, item_id, status, current_price, leader_user_id, start_time, end_time,
		       participant_count, viewer_count, increment_step, extension_seconds, extension_trigger_seconds,
		       ceiling_price, supports_auto_proxy
		FROM auction_sessions
		WHERE room_id = ?
		ORDER BY id
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]model.AuctionSession, 0)
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, *session)
	}
	return sessions, rows.Err()
}

func (r *MySQLSessionRepository) SaveSession(session model.AuctionSession) error {
	_, err := r.db.Exec(`
		INSERT INTO auction_sessions (id, room_id, item_id, status, current_price, leader_user_id, start_time, end_time, participant_count, viewer_count, increment_step, extension_seconds, extension_trigger_seconds, ceiling_price, supports_auto_proxy)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			room_id = VALUES(room_id),
			item_id = VALUES(item_id),
			status = VALUES(status),
			current_price = VALUES(current_price),
			leader_user_id = VALUES(leader_user_id),
			start_time = VALUES(start_time),
			end_time = VALUES(end_time),
			participant_count = VALUES(participant_count),
			viewer_count = VALUES(viewer_count),
			increment_step = VALUES(increment_step),
			extension_seconds = VALUES(extension_seconds),
			extension_trigger_seconds = VALUES(extension_trigger_seconds),
			ceiling_price = VALUES(ceiling_price),
			supports_auto_proxy = VALUES(supports_auto_proxy)
	`, session.ID, session.RoomID, session.ItemID, session.Status, session.CurrentPrice, session.LeaderUserID, nullableTimeValue(session.StartTime), nullableTimeValue(session.EndTime), session.ParticipantCount, session.ViewerCount, session.IncrementStep, session.ExtensionSeconds, session.ExtensionTrigger, nullableInt64Value(session.CeilingPrice), session.SupportsAutoProxy)
	return err
}

type sessionScanner interface {
	Scan(dest ...any) error
}

func scanSession(scanner sessionScanner) (*model.AuctionSession, error) {
	var session model.AuctionSession
	var leader sql.NullString
	var startTime sql.NullTime
	var endTime sql.NullTime
	var ceiling sql.NullInt64
	if err := scanner.Scan(
		&session.ID,
		&session.RoomID,
		&session.ItemID,
		&session.Status,
		&session.CurrentPrice,
		&leader,
		&startTime,
		&endTime,
		&session.ParticipantCount,
		&session.ViewerCount,
		&session.IncrementStep,
		&session.ExtensionSeconds,
		&session.ExtensionTrigger,
		&ceiling,
		&session.SupportsAutoProxy,
	); err != nil {
		return nil, err
	}
	session.LeaderUserID = nullableString(leader)
	session.StartTime = nullableTimeString(startTime)
	session.EndTime = nullableTimeString(endTime)
	session.CeilingPrice = nullableInt64Ptr(ceiling)
	return &session, nil
}
