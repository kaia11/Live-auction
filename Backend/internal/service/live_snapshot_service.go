package service

import (
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
	"auction-live/backend/internal/ws"
)

type LiveSnapshot struct {
	Room           model.LiveRoom           `json:"room"`
	CurrentSession model.AuctionSession     `json:"currentSession"`
	CurrentItem    model.AuctionItem        `json:"currentItem"`
	Items          []model.AuctionItem      `json:"items"`
	RankingTop3    []model.RankingEntry     `json:"rankingTop3"`
	MyStatus       *model.SessionUserStatus `json:"myStatus,omitempty"`
	LatestVersion  int64                    `json:"latestVersion"`
}

type LiveSnapshotService struct {
	store       *memoryStore
	hub         *ws.Hub
	runtime     *realtime.Runtime
	roomRepo    repository.RoomRepository
	itemRepo    repository.ItemRepository
	sessionRepo repository.SessionRepository
	bidRepo     repository.BidRepository
	userRepo    repository.UserRepository
}

func NewLiveSnapshotService(hub *ws.Hub, runtime *realtime.Runtime, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, sessionRepo repository.SessionRepository, bidRepo repository.BidRepository, userRepo repository.UserRepository) *LiveSnapshotService {
	return &LiveSnapshotService{
		store:       sharedStore,
		hub:         hub,
		runtime:     runtime,
		roomRepo:    roomRepo,
		itemRepo:    itemRepo,
		sessionRepo: sessionRepo,
		bidRepo:     bidRepo,
		userRepo:    userRepo,
	}
}

func (s *LiveSnapshotService) GetRoomSnapshot(roomID string, userID string) (LiveSnapshot, error) {
	if snapshot, ok, err := s.getRepositorySnapshot(roomID, userID); ok || err != nil {
		return snapshot, err
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	room, ok := s.store.rooms[roomID]
	if !ok {
		return LiveSnapshot{}, ErrRoomNotFound
	}

	currentSession := s.store.sessions[room.CurrentSessionID]
	currentItem := s.store.items[currentSession.ItemID]
	if s.runtime != nil {
		if state, ok, err := s.runtime.LoadSessionState(currentSession.ID); err == nil && ok {
			currentSession = overlaySessionState(currentSession, state)
		}
	}

	itemIDs := s.store.roomItems[roomID]
	items := make([]model.AuctionItem, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		items = append(items, s.store.items[itemID])
	}

	rankings := buildRankings(s.store.bids, s.store.users, currentSession.ID)
	if s.runtime != nil {
		if fromRedis, err := s.runtime.GetTopRanking(currentSession.ID, 3, s.store.users); err == nil && len(fromRedis) > 0 {
			rankings = fromRedis
		}
	}
	top3 := rankings
	if len(top3) > 3 {
		top3 = top3[:3]
	}

	var myStatus *model.SessionUserStatus
	if userID != "" {
		status := buildSessionUserStatus(currentSession, rankings, userID)
		if s.runtime != nil {
			if entry, ok, err := s.runtime.GetUserRankingEntry(currentSession.ID, userID, s.store.users); err == nil && ok {
				status.MyHighestBid = entry.HighestBid
				status.MyRank = entry.Rank
				status.IsLeading = entry.Rank == 1
			}
		}
		myStatus = &status
	}

	return LiveSnapshot{
		Room:           room,
		CurrentSession: currentSession,
		CurrentItem:    currentItem,
		Items:          items,
		RankingTop3:    top3,
		MyStatus:       myStatus,
		LatestVersion:  s.hub.LatestVersion(roomID),
	}, nil
}

func (s *LiveSnapshotService) getRepositorySnapshot(roomID string, userID string) (LiveSnapshot, bool, error) {
	if s.roomRepo == nil || s.itemRepo == nil || s.sessionRepo == nil {
		return LiveSnapshot{}, false, nil
	}

	room, err := s.roomRepo.GetRoomDetail(roomID)
	if err != nil || room == nil || room.ID == "" {
		return LiveSnapshot{}, false, nil
	}

	currentSession, err := s.sessionRepo.GetCurrentSession(roomID)
	if err != nil || currentSession == nil || currentSession.ID == "" {
		return LiveSnapshot{}, false, nil
	}

	currentItem, err := s.itemRepo.GetItemDetail(roomID, currentSession.ItemID)
	if err != nil || currentItem == nil || currentItem.ID == "" {
		return LiveSnapshot{}, false, nil
	}

	items, err := s.itemRepo.ListRoomItems(roomID)
	if err != nil {
		return LiveSnapshot{}, false, nil
	}

	sessionValue := *currentSession
	if s.runtime != nil {
		if state, ok, runtimeErr := s.runtime.LoadSessionState(currentSession.ID); runtimeErr == nil && ok {
			sessionValue = overlaySessionState(sessionValue, state)
		}
	}

	rankings := s.loadRankingEntries(currentSession.ID)
	top3 := rankings
	if len(top3) > 3 {
		top3 = top3[:3]
	}

	var myStatus *model.SessionUserStatus
	if userID != "" {
		status := buildSessionUserStatus(sessionValue, rankings, userID)
		if s.runtime != nil {
			if entry, ok, runtimeErr := s.runtime.GetUserRankingEntry(currentSession.ID, userID, s.store.users); runtimeErr == nil && ok {
				status.MyHighestBid = entry.HighestBid
				status.MyRank = entry.Rank
				status.IsLeading = entry.Rank == 1
			}
		}
		myStatus = &status
	}

	return LiveSnapshot{
		Room:           *room,
		CurrentSession: sessionValue,
		CurrentItem:    *currentItem,
		Items:          items,
		RankingTop3:    top3,
		MyStatus:       myStatus,
		LatestVersion:  s.hub.LatestVersion(roomID),
	}, true, nil
}

func (s *LiveSnapshotService) loadRankingEntries(sessionID string) []model.RankingEntry {
	if s.runtime != nil {
		if fromRedis, err := s.runtime.GetTopRanking(sessionID, 3, s.store.users); err == nil && len(fromRedis) > 0 {
			return fromRedis
		}
	}
	if s.bidRepo == nil {
		s.store.mu.RLock()
		defer s.store.mu.RUnlock()
		return buildRankings(s.store.bids, s.store.users, sessionID)
	}
	bids, err := s.bidRepo.ListSessionBids(sessionID)
	if err != nil || len(bids) == 0 {
		s.store.mu.RLock()
		defer s.store.mu.RUnlock()
		return buildRankings(s.store.bids, s.store.users, sessionID)
	}

	s.store.mu.RLock()
	users := make(map[string]model.User, len(s.store.users))
	for id, user := range s.store.users {
		users[id] = user
	}
	s.store.mu.RUnlock()
	if s.userRepo != nil {
		for _, bid := range bids {
			if _, ok := users[bid.UserID]; ok {
				continue
			}
			if user, userErr := s.userRepo.GetByID(bid.UserID); userErr == nil && user != nil && user.ID != "" {
				users[user.ID] = *user
			}
		}
	}
	return buildRankings(bids, users, sessionID)
}
