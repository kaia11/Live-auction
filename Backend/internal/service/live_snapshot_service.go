package service

import (
	"auction-live/backend/internal/model"
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
	store *memoryStore
	hub   *ws.Hub
}

func NewLiveSnapshotService(hub *ws.Hub) *LiveSnapshotService {
	return &LiveSnapshotService{store: sharedStore, hub: hub}
}

func (s *LiveSnapshotService) GetRoomSnapshot(roomID string, userID string) (LiveSnapshot, error) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	room, ok := s.store.rooms[roomID]
	if !ok {
		return LiveSnapshot{}, ErrRoomNotFound
	}

	currentSession := s.store.sessions[room.CurrentSessionID]
	currentItem := s.store.items[currentSession.ItemID]

	itemIDs := s.store.roomItems[roomID]
	items := make([]model.AuctionItem, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		items = append(items, s.store.items[itemID])
	}

	rankings := buildRankings(s.store.bids, s.store.users, currentSession.ID)
	top3 := rankings
	if len(top3) > 3 {
		top3 = top3[:3]
	}

	var myStatus *model.SessionUserStatus
	if userID != "" {
		status := buildSessionUserStatus(currentSession, rankings, userID)
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
