package service

import (
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/statemachine"
)

type OrderService struct {
	store *memoryStore
}

func NewOrderService() *OrderService {
	return &OrderService{store: sharedStore}
}

func (s *OrderService) ListMyOrders(userID string) []model.AuctionOrder {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	orders := make([]model.AuctionOrder, 0)
	for _, order := range s.store.orders {
		if order.BuyerUserID == userID {
			orders = append(orders, order)
		}
	}

	return orders
}

func (s *OrderService) UpdateOrderStatus(orderID string, event statemachine.OrderEvent) (model.AuctionOrder, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	order, ok := s.store.ordersByID[orderID]
	if !ok {
		return model.AuctionOrder{}, ErrOrderNotFound
	}

	nextStatus, err := statemachine.NextOrderState(order.Status, event)
	if err != nil {
		return model.AuctionOrder{}, ErrInvalidOrderState
	}

	order.Status = nextStatus
	s.store.ordersByID[orderID] = order
	s.store.ordersBySession[order.SessionID] = order
	for i := range s.store.orders {
		if s.store.orders[i].ID == order.ID {
			s.store.orders[i] = order
			break
		}
	}

	return order, nil
}
