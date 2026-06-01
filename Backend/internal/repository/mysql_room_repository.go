package repository

import (
	"database/sql"

	"auction-live/backend/internal/model"
)

type MySQLRoomRepository struct {
	db *sql.DB
}

func NewMySQLRoomRepository(db *sql.DB) *MySQLRoomRepository {
	return &MySQLRoomRepository{db: db}
}

func (r *MySQLRoomRepository) ListRooms() ([]model.LiveRoom, error) {
	rows, err := r.db.Query(`
		SELECT id, title, cover_image, video_url, status, anchor_user_id, anchor_name, online_count, thumbnail, current_session_id
		FROM live_rooms
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rooms := make([]model.LiveRoom, 0)
	for rows.Next() {
		var room model.LiveRoom
		var cover sql.NullString
		var video sql.NullString
		var thumbnail sql.NullString
		var currentSession sql.NullString
		if err := rows.Scan(
			&room.ID,
			&room.Title,
			&cover,
			&video,
			&room.Status,
			&room.AnchorUserID,
			&room.AnchorName,
			&room.OnlineCount,
			&thumbnail,
			&currentSession,
		); err != nil {
			return nil, err
		}
		room.CoverImage = nullableString(cover)
		room.VideoURL = nullableString(video)
		room.Thumbnail = nullableString(thumbnail)
		room.CurrentSessionID = nullableString(currentSession)
		rooms = append(rooms, room)
	}

	return rooms, rows.Err()
}

func (r *MySQLRoomRepository) GetRoomDetail(roomID string) (*model.LiveRoom, error) {
	row := r.db.QueryRow(`
		SELECT id, title, cover_image, video_url, status, anchor_user_id, anchor_name, online_count, thumbnail, current_session_id
		FROM live_rooms
		WHERE id = ?
	`, roomID)

	var room model.LiveRoom
	var cover sql.NullString
	var video sql.NullString
	var thumbnail sql.NullString
	var currentSession sql.NullString
	if err := row.Scan(
		&room.ID,
		&room.Title,
		&cover,
		&video,
		&room.Status,
		&room.AnchorUserID,
		&room.AnchorName,
		&room.OnlineCount,
		&thumbnail,
		&currentSession,
	); err != nil {
		return nil, err
	}

	room.CoverImage = nullableString(cover)
	room.VideoURL = nullableString(video)
	room.Thumbnail = nullableString(thumbnail)
	room.CurrentSessionID = nullableString(currentSession)
	return &room, nil
}

func (r *MySQLRoomRepository) SaveRoom(room model.LiveRoom) error {
	_, err := r.db.Exec(`
		INSERT INTO live_rooms (id, title, cover_image, video_url, status, anchor_user_id, anchor_name, online_count, thumbnail, current_session_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			cover_image = VALUES(cover_image),
			video_url = VALUES(video_url),
			status = VALUES(status),
			anchor_user_id = VALUES(anchor_user_id),
			anchor_name = VALUES(anchor_name),
			online_count = VALUES(online_count),
			thumbnail = VALUES(thumbnail),
			current_session_id = VALUES(current_session_id)
	`, room.ID, room.Title, nullableEmptyToNull(room.CoverImage), nullableEmptyToNull(room.VideoURL), room.Status, room.AnchorUserID, room.AnchorName, room.OnlineCount, nullableEmptyToNull(room.Thumbnail), nullableEmptyToNull(room.CurrentSessionID))
	return err
}
