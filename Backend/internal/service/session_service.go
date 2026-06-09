package service

import (
	"time"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
)

type SessionService struct {
	store    *memoryStore
	runtime  *realtime.Runtime
	repo     repository.SessionRepository
	bidRepo  repository.BidRepository
	userRepo repository.UserRepository
}

func NewSessionService(runtime *realtime.Runtime, repo repository.SessionRepository, bidRepo repository.BidRepository, userRepo repository.UserRepository) *SessionService {
	return &SessionService{
		store:    sharedStore,
		runtime:  runtime,
		repo:     repo,
		bidRepo:  bidRepo,
		userRepo: userRepo,
	}
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
	users := s.loadUsers()
	rankings := s.loadRankingEntries(sessionID, 3, users)

	top3 := rankings
	if len(top3) > 3 {
		top3 = top3[:3]
	}

	me := model.RankingEntry{}
	if s.runtime != nil {
		if entry, ok, err := s.runtime.GetUserRankingEntry(sessionID, "user-001", users); err == nil && ok {
			me = entry
		}
	}
	if me.UserID == "" {
		for _, entry := range s.loadRankingEntries(sessionID, 0, users) {
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
	session := s.loadSessionFromStore(sessionID)
	users := s.loadUsers()
	rankings := s.loadRankingEntries(sessionID, 100, users)
	if s.runtime != nil {
		if state, ok, err := s.runtime.LoadSessionState(sessionID); err == nil && ok {
			session = overlaySessionState(session, state)
		}
		if entry, ok, err := s.runtime.GetUserRankingEntry(sessionID, userID, users); err == nil && ok {
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

	status := buildSessionUserStatus(session, rankings, userID)
	if cfg, ok := s.loadAutoProxyConfig(sessionID, userID); ok {
		status.AutoProxyEnabled = true
		status.AutoProxyMaxPrice = cfg.MaxPrice
	}
	return status
}

func (s *SessionService) loadUsers() map[string]model.User {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	users := make(map[string]model.User, len(s.store.users))
	for id, user := range s.store.users {
		users[id] = user
	}
	return users
}

func (s *SessionService) loadSessionFromStore(sessionID string) model.AuctionSession {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	return s.store.sessions[sessionID]
}

func (s *SessionService) loadAutoProxyConfig(sessionID string, userID string) (model.AutoProxyConfig, bool) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	config, ok := s.store.autoProxyConfigs[sessionID+"::"+userID]
	return config, ok
}

func (s *SessionService) loadRankingEntries(sessionID string, limit int, users map[string]model.User) []model.RankingEntry {
	if s.runtime != nil {
		runtimeLimit := limit
		if runtimeLimit <= 0 {
			runtimeLimit = 100
		}
		if fromRedis, err := s.runtime.GetTopRanking(sessionID, runtimeLimit, users); err == nil && len(fromRedis) > 0 {
			return fromRedis
		}
	}

	if s.bidRepo != nil {
		bids, err := s.bidRepo.ListSessionBids(sessionID)
		if err == nil && len(bids) > 0 {
			hydrateUsers(users, s.userRepo, bids)
			rankings := buildRankings(bids, users, sessionID)
			if limit > 0 && len(rankings) > limit {
				return rankings[:limit]
			}
			return rankings
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	rankings := buildRankings(s.store.bids, s.store.users, sessionID)
	if limit > 0 && len(rankings) > limit {
		return rankings[:limit]
	}
	return rankings
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
