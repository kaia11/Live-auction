package service

import "auction-live/backend/internal/model"

type SessionService struct {
	store *memoryStore
}

func NewSessionService() *SessionService {
	return &SessionService{store: sharedStore}
}

func (s *SessionService) GetCurrentSession(roomID string) model.AuctionSession {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	room := s.store.rooms[roomID]
	return s.store.sessions[room.CurrentSessionID]
}

func (s *SessionService) GetRanking(sessionID string) map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	rankings := buildRankings(s.store.bids, s.store.users, sessionID)
	top3 := rankings
	if len(top3) > 3 {
		top3 = top3[:3]
	}

	me := model.RankingEntry{}
	for _, entry := range rankings {
		if entry.UserID == "user-001" {
			me = entry
			break
		}
	}

	return map[string]any{
		"sessionId": sessionID,
		"top3":      top3,
		"me":        me,
	}
}

func (s *SessionService) GetUserStatus(sessionID string, userID string) model.SessionUserStatus {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	session := s.store.sessions[sessionID]
	rankings := buildRankings(s.store.bids, s.store.users, sessionID)

	return buildSessionUserStatus(session, rankings, userID)
}

func buildSessionUserStatus(session model.AuctionSession, rankings []model.RankingEntry, userID string) model.SessionUserStatus {
	sessionID := session.ID

	status := model.SessionUserStatus{
		SessionID:      sessionID,
		UserID:         userID,
		CurrentPrice:   session.CurrentPrice,
		NextMinimumBid: session.CurrentPrice + session.IncrementStep,
	}

	for _, entry := range rankings {
		if entry.UserID == userID {
			status.MyHighestBid = entry.HighestBid
			status.MyRank = entry.Rank
			status.IsLeading = entry.Rank == 1
			break
		}
	}

	if status.IsLeading {
		status.VibrateSignalHint = "none"
	} else if status.MyHighestBid > 0 {
		status.VibrateSignalHint = "overtaken"
	} else {
		status.VibrateSignalHint = "none"
	}

	return status
}
