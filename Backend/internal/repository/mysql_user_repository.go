package repository

import (
	"database/sql"

	"auction-live/backend/internal/model"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
	return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) GetByID(userID string) (*model.User, error) {
	row := r.db.QueryRow(`
		SELECT id, nickname, avatar, role
		FROM users
		WHERE id = ?
	`, userID)

	var user model.User
	var avatar sql.NullString
	if err := row.Scan(&user.ID, &user.Nickname, &avatar, &user.Role); err != nil {
		return nil, err
	}
	user.Avatar = nullableString(avatar)
	return &user, nil
}
