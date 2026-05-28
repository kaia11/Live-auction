package statemachine

import "auction-live/backend/internal/domain"

type OrderEvent string

const (
	OrderEventMarkPaid OrderEvent = "mark_paid"
	OrderEventShip     OrderEvent = "ship"
	OrderEventComplete OrderEvent = "complete"
	OrderEventCancel   OrderEvent = "cancel"
)

func NextOrderState(current string, event OrderEvent) (string, error) {
	switch current {
	case domain.OrderStatusPendingPayment:
		switch event {
		case OrderEventMarkPaid:
			return domain.OrderStatusPaid, nil
		case OrderEventCancel:
			return domain.OrderStatusCancelled, nil
		}
	case domain.OrderStatusPaid:
		switch event {
		case OrderEventShip:
			return domain.OrderStatusShipped, nil
		case OrderEventCancel:
			return domain.OrderStatusCancelled, nil
		}
	case domain.OrderStatusShipped:
		if event == OrderEventComplete {
			return domain.OrderStatusCompleted, nil
		}
	}

	return "", ErrInvalidTransition
}
