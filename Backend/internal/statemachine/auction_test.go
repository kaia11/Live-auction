package statemachine

import (
	"errors"
	"testing"

	"auction-live/backend/internal/domain"
)

func TestNextSessionState(t *testing.T) {
	tests := []struct {
		name    string
		current string
		event   SessionEvent
		want    string
		wantErr error
	}{
		{name: "pending start", current: domain.SessionStatePending, event: SessionEventStart, want: domain.SessionStateBidding},
		{name: "bidding reach ceiling", current: domain.SessionStateBidding, event: SessionEventReachCeiling, want: domain.SessionStateEndedSold},
		{name: "bidding no bid timeout", current: domain.SessionStateBidding, event: SessionEventTimeoutNoBid, want: domain.SessionStateEndedPassed},
		{name: "pending cancel", current: domain.SessionStatePending, event: SessionEventCancel, want: domain.SessionStateCancelled},
		{name: "invalid sold start", current: domain.SessionStateEndedSold, event: SessionEventStart, wantErr: ErrInvalidTransition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextSessionState(tt.current, tt.event)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NextSessionState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NextSessionState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextQueueState(t *testing.T) {
	tests := []struct {
		name    string
		current string
		event   QueueEvent
		want    string
		wantErr error
	}{
		{name: "queued upcoming", current: domain.QueueStateQueued, event: QueueEventMarkUpcoming, want: domain.QueueStateUpcoming},
		{name: "upcoming activate", current: domain.QueueStateUpcoming, event: QueueEventActivate, want: domain.QueueStateActive},
		{name: "active finish", current: domain.QueueStateActive, event: QueueEventFinish, want: domain.QueueStateFinished},
		{name: "queued cancel", current: domain.QueueStateQueued, event: QueueEventCancel, want: domain.QueueStateCancelled},
		{name: "finished activate invalid", current: domain.QueueStateFinished, event: QueueEventActivate, wantErr: ErrInvalidTransition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextQueueState(tt.current, tt.event)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NextQueueState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NextQueueState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNextOrderState(t *testing.T) {
	tests := []struct {
		name    string
		current string
		event   OrderEvent
		want    string
		wantErr error
	}{
		{name: "pending paid", current: domain.OrderStatusPendingPayment, event: OrderEventMarkPaid, want: domain.OrderStatusPaid},
		{name: "paid shipped", current: domain.OrderStatusPaid, event: OrderEventShip, want: domain.OrderStatusShipped},
		{name: "shipped completed", current: domain.OrderStatusShipped, event: OrderEventComplete, want: domain.OrderStatusCompleted},
		{name: "pending cancelled", current: domain.OrderStatusPendingPayment, event: OrderEventCancel, want: domain.OrderStatusCancelled},
		{name: "completed ship invalid", current: domain.OrderStatusCompleted, event: OrderEventShip, wantErr: ErrInvalidTransition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NextOrderState(tt.current, tt.event)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NextOrderState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("NextOrderState() = %q, want %q", got, tt.want)
			}
		})
	}
}
