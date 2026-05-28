package service

import (
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/ws"
)

type SettlementScheduler struct {
	store    *memoryStore
	engine   *AuctionEngine
	hub      *ws.Hub
	interval time.Duration
	stopCh   chan struct{}
}

func NewSettlementScheduler(store *memoryStore, hub *ws.Hub, interval time.Duration) *SettlementScheduler {
	return &SettlementScheduler{
		store:    store,
		engine:   NewAuctionEngine(store),
		hub:      hub,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (s *SettlementScheduler) Start() {
	ticker := time.NewTicker(s.interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				outcomes := s.ScanOnce()
				for _, outcome := range outcomes {
					s.publishOutcome(outcome)
				}
			case <-s.stopCh:
				return
			}
		}
	}()
}

func (s *SettlementScheduler) Stop() {
	select {
	case <-s.stopCh:
		return
	default:
		close(s.stopCh)
	}
}

func (s *SettlementScheduler) ScanOnce() []SessionSettlement {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	now := time.Now()
	outcomes := make([]SessionSettlement, 0)
	for _, session := range s.store.sessions {
		if session.Status != domain.SessionStateBidding || session.EndTime == "" {
			continue
		}

		endTime, err := time.Parse(time.RFC3339, session.EndTime)
		if err != nil || endTime.After(now) {
			continue
		}

		outcome, err := s.engine.SettleSessionLocked(session.ID)
		if err != nil {
			logger.Error("scheduler settle failed session_id=%s error=%v", session.ID, err)
			continue
		}

		logger.Info("scheduler settled session session_id=%s room_id=%s status=%s", outcome.SessionID, outcome.RoomID, outcome.Status)
		outcomes = append(outcomes, outcome)
	}

	return outcomes
}

func (s *SettlementScheduler) publishOutcome(outcome SessionSettlement) {
	if outcome.RoomID == "" {
		return
	}

	s.hub.Publish(outcome.RoomID, ws.EventAuctionSessionEnded, outcome)
	if outcome.Order != nil {
		s.hub.Publish(outcome.RoomID, ws.EventAuctionOrderCreated, outcome.Order)
	}
	if outcome.NextSessionID != "" {
		s.hub.Publish(outcome.RoomID, ws.EventAuctionSessionUpcoming, map[string]any{
			"roomId":        outcome.RoomID,
			"nextSessionId": outcome.NextSessionID,
			"nextItemId":    outcome.NextItemID,
		})
	}
}
