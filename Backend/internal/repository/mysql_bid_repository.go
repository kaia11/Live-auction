package repository

import (
	"database/sql"

	"auction-live/backend/internal/model"
)

type MySQLBidRepository struct {
	db *sql.DB
}

func NewMySQLBidRepository(db *sql.DB) *MySQLBidRepository {
	return &MySQLBidRepository{db: db}
}

func (r *MySQLBidRepository) CreateBid(bid model.Bid) error {
	_, err := r.db.Exec(`
		INSERT INTO bids (id, session_id, room_id, item_id, user_id, bid_price, request_id, rank_after, status, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, bid.ID, bid.SessionID, bid.RoomID, bid.ItemID, bid.UserID, bid.BidPrice, bid.RequestID, bid.RankAfter, bid.Status, bid.CreateTime)
	return err
}

func (r *MySQLBidRepository) ListUserBids(userID string) ([]model.Bid, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, room_id, item_id, user_id, bid_price, request_id, rank_after, status, create_time
		FROM bids
		WHERE user_id = ?
		ORDER BY create_time DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bids := make([]model.Bid, 0)
	for rows.Next() {
		var bid model.Bid
		var createTime sql.NullTime
		if err := rows.Scan(
			&bid.ID,
			&bid.SessionID,
			&bid.RoomID,
			&bid.ItemID,
			&bid.UserID,
			&bid.BidPrice,
			&bid.RequestID,
			&bid.RankAfter,
			&bid.Status,
			&createTime,
		); err != nil {
			return nil, err
		}
		bid.CreateTime = nullableTimeString(createTime)
		bids = append(bids, bid)
	}
	return bids, rows.Err()
}

func (r *MySQLBidRepository) ListSessionBids(sessionID string) ([]model.Bid, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, room_id, item_id, user_id, bid_price, request_id, rank_after, status, create_time
		FROM bids
		WHERE session_id = ?
		ORDER BY create_time ASC
	`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bids := make([]model.Bid, 0)
	for rows.Next() {
		var bid model.Bid
		var createTime sql.NullTime
		if err := rows.Scan(
			&bid.ID,
			&bid.SessionID,
			&bid.RoomID,
			&bid.ItemID,
			&bid.UserID,
			&bid.BidPrice,
			&bid.RequestID,
			&bid.RankAfter,
			&bid.Status,
			&createTime,
		); err != nil {
			return nil, err
		}
		bid.CreateTime = nullableTimeString(createTime)
		bids = append(bids, bid)
	}
	return bids, rows.Err()
}
