package service

import (
	"time"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
)

type SessionService struct {
	store   *memoryStore
	runtime *realtime.Runtime
	repo    repository.SessionRepository
}

func NewSessionService(runtime *realtime.Runtime, repo repository.SessionRepository) *SessionService {
	return &SessionService{store: sharedStore, runtime: runtime, repo: repo}
}

func (s *SessionService) GetCurrentSession(roomID string) model.AuctionSession {
	if s.repo != nil {
		session, err := s.repo.GetCurrentSession(roomID)
		if err == nil && session != nil && session.ID != "" {
			if s.runtime == nil {
				return *session
			}

			state, ok, runtimeErr := s.runtime.LoadSessionState(session.ID)
			if runtimeErr != nil || !ok {
				return *session
			}
			return overlaySessionState(*session, state)
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	room := s.store.rooms[roomID]
	session := s.store.sessions[room.CurrentSessionID]
	if s.runtime == nil {
		return session
	}

	state, ok, err := s.runtime.LoadSessionState(session.ID)
	if err != nil || !ok {
		return session
	}
	return overlaySessionState(session, state)
}

func (s *SessionService) GetRanking(sessionID string) map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	rankings := buildRankings(s.store.bids, s.store.users, sessionID)
	if s.runtime != nil {
		if fromRedis, err := s.runtime.GetTopRanking(sessionID, 3, s.store.users); err == nil {
			rankings = fromRedis
		}
	}

	top3 := rankings
	if len(top3) > 3 {
		top3 = top3[:3]
	}

	me := model.RankingEntry{}
	if s.runtime != nil {
		if entry, ok, err := s.runtime.GetUserRankingEntry(sessionID, "user-001", s.store.users); err == nil && ok {
			me = entry
		}
	}
	if me.UserID == "" {
		for _, entry := range buildRankings(s.store.bids, s.store.users, sessionID) {
			if entry.UserID == "user-001" {
				me = entry
				break
			}
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
	if s.runtime != nil {
		if state, ok, err := s.runtime.LoadSessionState(sessionID); err == nil && ok {
			session = overlaySessionState(session, state)
		}
		if top, err := s.runtime.GetTopRanking(sessionID, 100, s.store.users); err == nil && len(top) > 0 {
			rankings = top
			if entry, ok, err := s.runtime.GetUserRankingEntry(sessionID, userID, s.store.users); err == nil && ok {
				found := false
				for _, item := range rankings {
					if item.UserID == userID {
						found = true
						break
					}
				}
				if !found {
					rankings = append(rankings, entry)
				}
			}
		}
	}

	return buildSessionUserStatus(session, rankings, userID)
}

func overlaySessionState(session model.AuctionSession, state realtime.SessionState) model.AuctionSession {
	session.Status = state.Status
	session.CurrentPrice = state.CurrentPrice
	session.LeaderUserID = state.LeaderUserID
	session.ParticipantCount = state.ParticipantCount
	session.IncrementStep = state.IncrementStep
	session.ExtensionSeconds = state.ExtensionSeconds
	session.ExtensionTrigger = state.ExtensionTriggerSeconds
	session.CeilingPrice = state.CeilingPrice
	if state.EndTimeUnix > 0 {
		session.EndTime = time.Unix(state.EndTimeUnix, 0).Format(time.RFC3339)
	} else {
		session.EndTime = ""
	}
	return session
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
