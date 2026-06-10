package service

import (
	"testing"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

func TestOrderServicePayOrderMarksPendingPaymentAsPaid(t *testing.T) {
	store := newMemoryStore()
	order := model.AuctionOrder{
		ID:          "order-001",
		SessionID:   "session-001",
		RoomID:      "room-001",
		ItemID:      "item-001",
		BuyerUserID: "user-001",
		Amount:      188,
		Status:      domain.OrderStatusPendingPayment,
		CreateTime:  "2026-06-10T10:00:00Z",
	}
	store.orders = append(store.orders, order)
	store.ordersByID[order.ID] = order
	store.ordersBySession[order.SessionID] = order

	service := &OrderService{
		store: store,
		repo:  repository.NewMemoryOrderRepository(store),
	}

	updated, err := service.PayOrder(order.ID, "user-001")
	if err != nil {
		t.Fatalf("PayOrder() error = %v", err)
	}
	if updated.Status != domain.OrderStatusPaid {
		t.Fatalf("expected paid status, got %q", updated.Status)
	}
	if got := store.ordersByID[order.ID].Status; got != domain.OrderStatusPaid {
		t.Fatalf("expected store status paid, got %q", got)
	}
}

func TestOrderServicePayOrderRejectsNonBuyer(t *testing.T) {
	store := newMemoryStore()
	order := model.AuctionOrder{
		ID:          "order-002",
		SessionID:   "session-001",
		RoomID:      "room-001",
		ItemID:      "item-001",
		BuyerUserID: "user-001",
		Amount:      188,
		Status:      domain.OrderStatusPendingPayment,
		CreateTime:  "2026-06-10T10:00:00Z",
	}
	store.orders = append(store.orders, order)
	store.ordersByID[order.ID] = order
	store.ordersBySession[order.SessionID] = order

	service := &OrderService{
		store: store,
		repo:  repository.NewMemoryOrderRepository(store),
	}

	if _, err := service.PayOrder(order.ID, "user-002"); err != ErrOrderNotFound {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}
