package service

import (
	"sort"
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
)

type BidService struct {
	store  *memoryStore
	engine *AuctionEngine
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
	return &BidService{store: sharedStore, engine: NewAuctionEngine(sharedStore)}
}

func (s *BidService) CreateBid(input CreateBidInput) (model.BidResult, *SessionSettlement, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	if _, ok := s.store.users[input.UserID]; !ok {
		return model.BidResult{}, nil, ErrUserNotFound
	}

	room, ok := s.store.rooms[input.RoomID]
	if !ok {
		return model.BidResult{}, nil, ErrRoomNotFound
	}

	session, ok := s.store.sessions[input.SessionID]
	if !ok {
		return model.BidResult{}, nil, ErrSessionNotFound
	}

	item, ok := s.store.items[input.ItemID]
	if !ok {
		return model.BidResult{}, nil, ErrItemNotFound
	}

	if room.CurrentSessionID != session.ID || session.RoomID != room.ID || session.ItemID != item.ID || item.RoomID != room.ID {
		return model.BidResult{}, nil, ErrBidOwnershipMismatch
	}

	if session.Status != domain.SessionStateBidding {
		return model.BidResult{}, nil, ErrSessionNotBidding
	}

	if cached, exists := s.store.processedRequests[input.RequestID]; exists {
		return cached, nil, ErrDuplicateBidRequest
	}

	nextMinimumBid := session.CurrentPrice + session.IncrementStep
	if session.CurrentPrice == 0 && input.BidPrice < item.StartPrice {
		return model.BidResult{}, nil, ErrInvalidBidPrice
	}

	if input.BidPrice < nextMinimumBid {
		return model.BidResult{}, nil, ErrInvalidBidPrice
	}

	if session.IncrementStep > 0 {
		delta := input.BidPrice - session.CurrentPrice
		if delta%session.IncrementStep != 0 {
			return model.BidResult{}, nil, ErrInvalidBidPrice
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
		Status:     domain.BidStatusAccepted,
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

	var settlement *SessionSettlement
	if ceilingReached {
		outcome, err := s.engine.ReachCeilingLocked(session.ID)
		if err != nil {
			return model.BidResult{}, nil, err
		}
		settlement = &outcome
	} else {
		s.store.sessions[session.ID] = session
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

	return result, settlement, nil
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

func (s *BidService) ListMyBidHistories(userID string) []model.UserBidHistory {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	histories := make([]model.UserBidHistory, 0)
	for _, bid := range s.store.bids {
		if bid.UserID != userID {
			continue
		}

		item := s.store.items[bid.ItemID]
		session := s.store.sessions[bid.SessionID]

		histories = append(histories, model.UserBidHistory{
			ID:        bid.ID,
			ItemID:    bid.ItemID,
			ItemTitle: item.Title,
			ItemImage: item.CoverImage,
			BidPrice:  bid.BidPrice,
			Result:    resolveBidHistoryResult(session, userID),
			BidTime:   bid.CreateTime,
		})
	}

	sort.SliceStable(histories, func(i, j int) bool {
		return histories[i].BidTime > histories[j].BidTime
	})

	return histories
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

func resolveBidHistoryResult(session model.AuctionSession, userID string) string {
	switch session.Status {
	case domain.SessionStateBidding, domain.SessionStatePending:
		return "pending"
	case domain.SessionStateEndedSold:
		if session.LeaderUserID == userID {
			return "win"
		}
		return "lose"
	case domain.SessionStateEndedPassed, domain.SessionStateCancelled:
		return "lose"
	default:
		if session.LeaderUserID == userID {
			return "win"
		}
		return "pending"
	}
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
