package repository

import (
	"database/sql"
	"fmt"

	"auction-live/backend/internal/model"
)

type MySQLItemRepository struct {
	db *sql.DB
}

func NewMySQLItemRepository(db *sql.DB) *MySQLItemRepository {
	return &MySQLItemRepository{db: db}
}

func (r *MySQLItemRepository) ListRoomItems(roomID string) ([]model.AuctionItem, error) {
	rows, err := r.db.Query(`
		SELECT id, room_id, title, cover_image, description, start_price, increment_step, ceiling_price,
		       duration_seconds, extension_seconds, extension_trigger_seconds, queue_status
		FROM auction_items
		WHERE room_id = ?
		ORDER BY COALESCE((SELECT sort_order FROM room_item_queue q WHERE q.room_id = auction_items.room_id AND q.item_id = auction_items.id), 999999), id
	`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.AuctionItem, 0)
	for rows.Next() {
		var item model.AuctionItem
		var ceiling sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.RoomID,
			&item.Title,
			&item.CoverImage,
			&item.Description,
			&item.StartPrice,
			&item.IncrementStep,
			&ceiling,
			&item.DurationSeconds,
			&item.ExtensionSeconds,
			&item.ExtensionTriggerSeconds,
			&item.QueueStatus,
		); err != nil {
			return nil, err
		}
		item.CeilingPrice = nullableInt64Ptr(ceiling)
		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *MySQLItemRepository) GetItemDetail(roomID string, itemID string) (*model.AuctionItem, error) {
	row := r.db.QueryRow(`
		SELECT id, room_id, title, cover_image, description, start_price, increment_step, ceiling_price,
		       duration_seconds, extension_seconds, extension_trigger_seconds, queue_status
		FROM auction_items
		WHERE room_id = ? AND id = ?
	`, roomID, itemID)

	var item model.AuctionItem
	var ceiling sql.NullInt64
	if err := row.Scan(
		&item.ID,
		&item.RoomID,
		&item.Title,
		&item.CoverImage,
		&item.Description,
		&item.StartPrice,
		&item.IncrementStep,
		&ceiling,
		&item.DurationSeconds,
		&item.ExtensionSeconds,
		&item.ExtensionTriggerSeconds,
		&item.QueueStatus,
	); err != nil {
		return nil, err
	}
	item.CeilingPrice = nullableInt64Ptr(ceiling)
	return &item, nil
}

func (r *MySQLItemRepository) SaveItem(item model.AuctionItem) error {
	_, err := r.db.Exec(`
		INSERT INTO auction_items (id, room_id, title, cover_image, description, start_price, increment_step, ceiling_price, duration_seconds, extension_seconds, extension_trigger_seconds, queue_status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			title = VALUES(title),
			cover_image = VALUES(cover_image),
			description = VALUES(description),
			start_price = VALUES(start_price),
			increment_step = VALUES(increment_step),
			ceiling_price = VALUES(ceiling_price),
			duration_seconds = VALUES(duration_seconds),
			extension_seconds = VALUES(extension_seconds),
			extension_trigger_seconds = VALUES(extension_trigger_seconds),
			queue_status = VALUES(queue_status)
	`, item.ID, item.RoomID, item.Title, item.CoverImage, item.Description, item.StartPrice, item.IncrementStep, nullableInt64Value(item.CeilingPrice), item.DurationSeconds, item.ExtensionSeconds, item.ExtensionTriggerSeconds, item.QueueStatus)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT IGNORE INTO room_item_queue (room_id, item_id, sort_order)
		VALUES (?, ?, ?)
	`, item.RoomID, item.ID, 999999)
	return err
}

func (r *MySQLItemRepository) ReplaceRoomQueue(roomID string, itemIDs []string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM room_item_queue WHERE room_id = ?`, roomID); err != nil {
		return err
	}

	for index, itemID := range itemIDs {
		if _, err := tx.Exec(`
			INSERT INTO room_item_queue (room_id, item_id, sort_order)
			VALUES (?, ?, ?)
		`, roomID, itemID, index+1); err != nil {
			return fmt.Errorf("replace room queue failed for item %s: %w", itemID, err)
		}
	}

	return tx.Commit()
}
