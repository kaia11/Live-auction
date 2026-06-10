package service

import (
	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
	"auction-live/backend/internal/statemachine"
)

type OrderService struct {
	store *memoryStore
	repo  repository.OrderRepository
}

func NewOrderService(repo repository.OrderRepository) *OrderService {
	return &OrderService{store: sharedStore, repo: repo}
}

func (s *OrderService) ListMyOrders(userID string) []model.AuctionOrder {
	if s.repo != nil {
		orders, err := s.repo.ListUserOrders(userID)
		if err == nil && len(orders) > 0 {
			return orders
		}
	}

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

func (s *OrderService) PayOrder(orderID string, userID string) (model.AuctionOrder, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	order, err := s.loadOrderLocked(orderID)
	if err != nil {
		return model.AuctionOrder{}, err
	}
	if order.BuyerUserID != userID {
		return model.AuctionOrder{}, ErrOrderNotFound
	}

	return s.applyOrderEventLocked(orderID, statemachine.OrderEventMarkPaid)
}

func (s *OrderService) UpdateOrderStatus(orderID string, event statemachine.OrderEvent) (model.AuctionOrder, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	return s.applyOrderEventLocked(orderID, event)
}

func (s *OrderService) loadOrderLocked(orderID string) (model.AuctionOrder, error) {
	order, ok := s.store.ordersByID[orderID]
	if ok {
		return order, nil
	}

	if s.repo == nil {
		return model.AuctionOrder{}, ErrOrderNotFound
	}
	persistedOrder, err := s.repo.GetOrderByID(orderID)
	if err != nil || persistedOrder == nil || persistedOrder.ID == "" {
		return model.AuctionOrder{}, ErrOrderNotFound
	}

	order = *persistedOrder
	s.store.ordersByID[orderID] = order
	s.store.ordersBySession[order.SessionID] = order
	found := false
	for i := range s.store.orders {
		if s.store.orders[i].ID == order.ID {
			s.store.orders[i] = order
			found = true
			break
		}
	}
	if !found {
		s.store.orders = append(s.store.orders, order)
	}

	return order, nil
}

func (s *OrderService) applyOrderEventLocked(orderID string, event statemachine.OrderEvent) (model.AuctionOrder, error) {
	order, err := s.loadOrderLocked(orderID)
	if err != nil {
		return model.AuctionOrder{}, err
	}

	if event == statemachine.OrderEventMarkPaid && order.Status != domain.OrderStatusPendingPayment {
		return model.AuctionOrder{}, ErrInvalidOrderState
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
	if s.repo != nil {
		_ = s.repo.UpdateOrder(order)
	}

	return order, nil
}
