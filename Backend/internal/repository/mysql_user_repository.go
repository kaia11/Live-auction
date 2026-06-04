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
		SELECT id, username, password, nickname, avatar, role
		FROM users
		WHERE id = ?
	`, userID)

	var user model.User
	var avatar sql.NullString
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Nickname, &avatar, &user.Role); err != nil {
		return nil, err
	}
	user.Avatar = nullableString(avatar)
	return &user, nil
}

func (r *MySQLUserRepository) GetByUsername(username string) (*model.User, error) {
	row := r.db.QueryRow(`
		SELECT id, username, password, nickname, avatar, role
		FROM users
		WHERE username = ?
	`, username)

	var user model.User
	var avatar sql.NullString
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Nickname, &avatar, &user.Role); err != nil {
		return nil, err
	}
	user.Avatar = nullableString(avatar)
	return &user, nil
}

func (r *MySQLUserRepository) Create(user model.User) error {
	_, err := r.db.Exec(`
		INSERT INTO users (id, username, password, nickname, avatar, role)
		VALUES (?, ?, ?, ?, ?, ?)
	`, user.ID, user.Username, user.Password, user.Nickname, nullableEmptyToNull(user.Avatar), user.Role)
	return err
}

func (r *MySQLUserRepository) UpdatePasswordHash(userID string, passwordHash string) error {
	_, err := r.db.Exec(`
		UPDATE users
		SET password = ?
		WHERE id = ?
	`, passwordHash, userID)
	return err
}

func (r *MySQLUserRepository) List() ([]model.User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, password, nickname, avatar, role
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		var avatar sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &user.Password, &user.Nickname, &avatar, &user.Role); err != nil {
			return nil, err
		}
		user.Avatar = nullableString(avatar)
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
