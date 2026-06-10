package repository

import (
	"database/sql"

	"auction-live/backend/internal/model"
)

type MySQLOrderRepository struct {
	db *sql.DB
}

func NewMySQLOrderRepository(db *sql.DB) *MySQLOrderRepository {
	return &MySQLOrderRepository{db: db}
}

func (r *MySQLOrderRepository) CreateOrder(order model.AuctionOrder) error {
	_, err := r.db.Exec(`
		INSERT INTO auction_orders (id, session_id, room_id, item_id, buyer_user_id, amount, status, create_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			buyer_user_id = VALUES(buyer_user_id),
			amount = VALUES(amount),
			status = VALUES(status),
			create_time = VALUES(create_time)
	`, order.ID, order.SessionID, order.RoomID, order.ItemID, order.BuyerUserID, order.Amount, order.Status, nullableTimeValue(order.CreateTime))
	return err
}

func (r *MySQLOrderRepository) GetOrderByID(orderID string) (*model.AuctionOrder, error) {
	row := r.db.QueryRow(`
		SELECT id, session_id, room_id, item_id, buyer_user_id, amount, status, create_time
		FROM auction_orders
		WHERE id = ?
	`, orderID)

	var order model.AuctionOrder
	var createTime sql.NullTime
	if err := row.Scan(
		&order.ID,
		&order.SessionID,
		&order.RoomID,
		&order.ItemID,
		&order.BuyerUserID,
		&order.Amount,
		&order.Status,
		&createTime,
	); err != nil {
		return nil, err
	}
	order.CreateTime = nullableTimeString(createTime)
	return &order, nil
}

func (r *MySQLOrderRepository) ListAllOrders() ([]model.AuctionOrder, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, room_id, item_id, buyer_user_id, amount, status, create_time
		FROM auction_orders
		ORDER BY create_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.AuctionOrder, 0)
	for rows.Next() {
		var order model.AuctionOrder
		var createTime sql.NullTime
		if err := rows.Scan(
			&order.ID,
			&order.SessionID,
			&order.RoomID,
			&order.ItemID,
			&order.BuyerUserID,
			&order.Amount,
			&order.Status,
			&createTime,
		); err != nil {
			return nil, err
		}
		order.CreateTime = nullableTimeString(createTime)
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (r *MySQLOrderRepository) ListUserOrders(userID string) ([]model.AuctionOrder, error) {
	rows, err := r.db.Query(`
		SELECT id, session_id, room_id, item_id, buyer_user_id, amount, status, create_time
		FROM auction_orders
		WHERE buyer_user_id = ?
		ORDER BY create_time DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]model.AuctionOrder, 0)
	for rows.Next() {
		var order model.AuctionOrder
		var createTime sql.NullTime
		if err := rows.Scan(
			&order.ID,
			&order.SessionID,
			&order.RoomID,
			&order.ItemID,
			&order.BuyerUserID,
			&order.Amount,
			&order.Status,
			&createTime,
		); err != nil {
			return nil, err
		}
		order.CreateTime = nullableTimeString(createTime)
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (r *MySQLOrderRepository) UpdateOrder(order model.AuctionOrder) error {
	_, err := r.db.Exec(`
		UPDATE auction_orders
		SET status = ?, amount = ?, buyer_user_id = ?, room_id = ?, item_id = ?, create_time = ?
		WHERE id = ?
	`, order.Status, order.Amount, order.BuyerUserID, order.RoomID, order.ItemID, nullableTimeValue(order.CreateTime), order.ID)
	return err
}
