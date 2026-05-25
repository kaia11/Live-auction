package service

import (
	"sort"
	"time"

	"auction-live/backend/internal/model"
)

type BidService struct {
	store *memoryStore
}

type CreateBidInput struct {
	RoomID    string
	SessionID string
	ItemID    string
	UserID    string
	BidPrice  int64
	RequestID string
}

func NewBidService() *BidService {
	return &BidService{store: sharedStore}
}

func (s *BidService) CreateBid(input CreateBidInput) (model.BidResult, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	if _, ok := s.store.users[input.UserID]; !ok {
		return model.BidResult{}, ErrUserNotFound
	}

	room, ok := s.store.rooms[input.RoomID]
	if !ok {
		return model.BidResult{}, ErrRoomNotFound
	}

	session, ok := s.store.sessions[input.SessionID]
	if !ok {
		return model.BidResult{}, ErrSessionNotFound
	}

	item, ok := s.store.items[input.ItemID]
	if !ok {
		return model.BidResult{}, ErrItemNotFound
	}

	if room.CurrentSessionID != session.ID || session.RoomID != room.ID || session.ItemID != item.ID || item.RoomID != room.ID {
		return model.BidResult{}, ErrBidOwnershipMismatch
	}

	if session.Status != "bidding" {
		return model.BidResult{}, ErrSessionNotBidding
	}

	if cached, exists := s.store.processedRequests[input.RequestID]; exists {
		return cached, ErrDuplicateBidRequest
	}

	nextMinimumBid := session.CurrentPrice + session.IncrementStep
	if session.CurrentPrice == 0 && input.BidPrice < item.StartPrice {
		return model.BidResult{}, ErrInvalidBidPrice
	}

	if input.BidPrice < nextMinimumBid {
		return model.BidResult{}, ErrInvalidBidPrice
	}

	if session.IncrementStep > 0 {
		delta := input.BidPrice - session.CurrentPrice
		if delta%session.IncrementStep != 0 {
			return model.BidResult{}, ErrInvalidBidPrice
		}
	}

	acceptedBidPrice := input.BidPrice
	ceilingReached := false
	if session.CeilingPrice != nil && input.BidPrice >= *session.CeilingPrice {
		acceptedBidPrice = *session.CeilingPrice
		ceilingReached = true
	}

	session.CurrentPrice = acceptedBidPrice
	session.LeaderUserID = input.UserID
	session.ParticipantCount = currentParticipantsCount(s.store.bids, session.ID, input.UserID)

	extensionApplied := false
	if !ceilingReached {
		endTime, err := time.Parse(time.RFC3339, session.EndTime)
		if err == nil {
			if endTime.Sub(time.Now()) <= time.Duration(session.ExtensionTrigger)*time.Second {
				session.EndTime = endTime.Add(time.Duration(session.ExtensionSeconds) * time.Second).Format(time.RFC3339)
				extensionApplied = true
			}
		}
	}

	bid := model.Bid{
		ID:         nextBidID(len(s.store.bids)),
		SessionID:  input.SessionID,
		RoomID:     input.RoomID,
		ItemID:     input.ItemID,
		UserID:     input.UserID,
		BidPrice:   acceptedBidPrice,
		RequestID:  input.RequestID,
		Status:     "accepted",
		CreateTime: time.Now().Format(time.RFC3339),
	}
	s.store.bids = append(s.store.bids, bid)

	rankings := buildRankings(s.store.bids, s.store.users, input.SessionID)
	for _, entry := range rankings {
		if entry.UserID == input.UserID {
			bid.RankAfter = entry.Rank
			break
		}
	}
	s.store.bids[len(s.store.bids)-1] = bid

	if ceilingReached {
		session.Status = "ended_sold"
	}
	s.store.sessions[session.ID] = session

	if ceilingReached {
		s.finishCurrentSessionLocked(session, "ended_sold")
	}

	result := model.BidResult{
		RoomID:            input.RoomID,
		SessionID:         input.SessionID,
		ItemID:            input.ItemID,
		UserID:            input.UserID,
		AcceptedBidPrice:  acceptedBidPrice,
		RequestID:         input.RequestID,
		CurrentPrice:      acceptedBidPrice,
		IsLeading:         true,
		ExtensionApplied:  extensionApplied,
		CeilingReached:    ceilingReached,
		NextMinimumBid:    acceptedBidPrice + session.IncrementStep,
		VibrateSignalHint: "overtake",
	}
	s.store.processedRequests[input.RequestID] = result

	return result, nil
}

func (s *BidService) ListMyBids(userID string) []model.Bid {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	bids := make([]model.Bid, 0)
	for _, bid := range s.store.bids {
		if bid.UserID == userID {
			bids = append(bids, bid)
		}
	}

	sort.SliceStable(bids, func(i, j int) bool {
		return bids[i].CreateTime > bids[j].CreateTime
	})

	return bids
}

func currentParticipantsCount(bids []model.Bid, sessionID string, userID string) int {
	seen := map[string]struct{}{userID: {}}
	for _, bid := range bids {
		if bid.SessionID == sessionID {
			seen[bid.UserID] = struct{}{}
		}
	}

	return len(seen)
}

func buildRankings(bids []model.Bid, users map[string]model.User, sessionID string) []model.RankingEntry {
	highestByUser := make(map[string]int64)
	for _, bid := range bids {
		if bid.SessionID != sessionID {
			continue
		}
		if bid.BidPrice > highestByUser[bid.UserID] {
			highestByUser[bid.UserID] = bid.BidPrice
		}
	}

	rankings := make([]model.RankingEntry, 0, len(highestByUser))
	for userID, highestBid := range highestByUser {
		user := users[userID]
		rankings = append(rankings, model.RankingEntry{
			UserID:     userID,
			Nickname:   user.Nickname,
			Avatar:     user.Avatar,
			HighestBid: highestBid,
		})
	}

	sort.SliceStable(rankings, func(i, j int) bool {
		return rankings[i].HighestBid > rankings[j].HighestBid
	})

	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	return rankings
}

func (s *BidService) finishCurrentSessionLocked(session model.AuctionSession, finalStatus string) {
	session.Status = finalStatus
	s.store.sessions[session.ID] = session

	item := s.store.items[session.ItemID]
	item.QueueStatus = "finished"
	s.store.items[item.ID] = item

	nextSessionID, nextItemID := s.findNextQueuedLocked(session.RoomID)
	if nextSessionID == "" || nextItemID == "" {
		room := s.store.rooms[session.RoomID]
		room.CurrentSessionID = session.ID
		s.store.rooms[session.RoomID] = room
		return
	}

	nextItem := s.store.items[nextItemID]
	nextItem.QueueStatus = "upcoming"
	s.store.items[nextItemID] = nextItem

	nextSession := s.store.sessions[nextSessionID]
	nextSession.Status = "pending"
	nextSession.CurrentPrice = nextItem.StartPrice
	nextSession.LeaderUserID = ""
	nextSession.ParticipantCount = 0
	nextSession.EndTime = ""
	nextSession.IncrementStep = nextItem.IncrementStep
	nextSession.ExtensionSeconds = nextItem.ExtensionSeconds
	nextSession.ExtensionTrigger = nextItem.ExtensionTriggerSeconds
	nextSession.CeilingPrice = nextItem.CeilingPrice
	s.store.sessions[nextSessionID] = nextSession

	room := s.store.rooms[session.RoomID]
	room.CurrentSessionID = nextSessionID
	s.store.rooms[session.RoomID] = room
}

func (s *BidService) findNextQueuedLocked(roomID string) (string, string) {
	for _, itemID := range s.store.roomItems[roomID] {
		item := s.store.items[itemID]
		if item.QueueStatus == "queued" || item.QueueStatus == "upcoming" {
			for sessionID, session := range s.store.sessions {
				if session.RoomID == roomID && session.ItemID == itemID {
					return sessionID, itemID
				}
			}
		}
	}
	return "", ""
}
