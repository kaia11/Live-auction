package repository

import (
	"database/sql"

	"auction-live/backend/internal/model"
)

type MySQLCommentRepository struct {
	db *sql.DB
}

func NewMySQLCommentRepository(db *sql.DB) *MySQLCommentRepository {
	return &MySQLCommentRepository{db: db}
}

func (r *MySQLCommentRepository) CreateComment(comment model.RoomComment) error {
	_, err := r.db.Exec(`
		INSERT INTO room_comments (room_id, user_id, nickname, content, create_time)
		VALUES (?, ?, ?, ?, ?)
	`, comment.RoomID, comment.UserID, comment.Nickname, comment.Content, nullableTimeValue(comment.CreateTime))
	return err
}
