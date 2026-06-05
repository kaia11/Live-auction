package service

import (
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/monitoring"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
	"auction-live/backend/internal/ws"
)

type SettlementScheduler struct {
	store    *memoryStore
	engine   *AuctionEngine
	hub      *ws.Hub
	runtime  *realtime.Runtime
	metrics  *monitoring.Metrics
	interval time.Duration
	stopCh   chan struct{}
	leaseID  string
}

func NewSettlementScheduler(store *memoryStore, hub *ws.Hub, runtime *realtime.Runtime, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, sessionRepo repository.SessionRepository, resultRepo repository.ResultRepository, orderRepo repository.OrderRepository, metrics *monitoring.Metrics, interval time.Duration) *SettlementScheduler {
	leaseID := "local-scheduler"
	if runtime != nil {
		leaseID = runtime.NewLeaseOwner("settlement")
	}
	return &SettlementScheduler{
		store:    store,
		engine:   NewAuctionEngine(store, runtime, roomRepo, itemRepo, sessionRepo, resultRepo, orderRepo),
		hub:      hub,
		runtime:  runtime,
		metrics:  metrics,
		interval: interval,
		stopCh:   make(chan struct{}),
		leaseID:  leaseID,
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
	sessions := s.snapshotSessions()
	now := time.Now()
	outcomes := make([]SessionSettlement, 0)
	for _, localSession := range sessions {
		session := localSession
		if s.runtime != nil {
			if state, ok, err := s.runtime.LoadSessionState(session.ID); err == nil && ok {
				session = overlaySessionState(session, state)
			}
		}

		if session.Status != domain.SessionStateBidding || session.EndTime == "" {
			continue
		}

		endTime, err := time.Parse(time.RFC3339, session.EndTime)
		if err != nil || endTime.After(now) {
			continue
		}

		if !s.tryClaimSettlement(session.ID) {
			continue
		}
		delayMS := now.Sub(endTime).Milliseconds()

		outcome, err := s.settleSession(session)
		if err != nil {
			logger.Error("scheduler settle failed session_id=%s error=%v", session.ID, err)
			if s.metrics != nil {
				s.metrics.RecordError("settlement_exception")
			}
			continue
		}
		if s.metrics != nil {
			s.metrics.RecordSettlement(delayMS)
		}

		logger.Info("scheduler settled session session_id=%s room_id=%s status=%s", outcome.SessionID, outcome.RoomID, outcome.Status)
		outcomes = append(outcomes, outcome)
	}

	return outcomes
}

func (s *SettlementScheduler) snapshotSessions() []model.AuctionSession {
	if sessions, ok := s.snapshotSessionsFromRepositories(); ok {
		return sessions
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	sessions := make([]model.AuctionSession, 0, len(s.store.sessions))
	for _, session := range s.store.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *SettlementScheduler) snapshotSessionsFromRepositories() ([]model.AuctionSession, bool) {
	if s.engine == nil || s.engine.roomRepo == nil || s.engine.sessionRepo == nil {
		return nil, false
	}

	rooms, err := s.engine.roomRepo.ListRooms()
	if err != nil || len(rooms) == 0 {
		return nil, false
	}

	sessions := make([]model.AuctionSession, 0)
	for _, room := range rooms {
		roomSessions, err := s.engine.sessionRepo.ListRoomSessions(room.ID)
		if err != nil {
			return nil, false
		}
		sessions = append(sessions, roomSessions...)
	}

	return sessions, true
}

func (s *SettlementScheduler) tryClaimSettlement(sessionID string) bool {
	if s.runtime == nil {
		return true
	}
	ok, err := s.runtime.TryAcquireSettlementLease(sessionID, s.leaseID, 5*time.Second)
	if err != nil {
		logger.Error("scheduler lease failed session_id=%s error=%v", sessionID, err)
		if s.metrics != nil {
			s.metrics.RecordError("settlement_lease_error")
		}
		return false
	}
	return ok
}

func (s *SettlementScheduler) settleSession(session model.AuctionSession) (SessionSettlement, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	if current, ok := s.store.sessions[session.ID]; ok {
		current.Status = session.Status
		current.CurrentPrice = session.CurrentPrice
		current.LeaderUserID = session.LeaderUserID
		current.EndTime = session.EndTime
		current.ParticipantCount = session.ParticipantCount
		current.IncrementStep = session.IncrementStep
		current.ExtensionSeconds = session.ExtensionSeconds
		current.ExtensionTrigger = session.ExtensionTrigger
		current.CeilingPrice = session.CeilingPrice
		s.store.sessions[session.ID] = current
	}

	return s.engine.SettleSessionLocked(session.ID)
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
