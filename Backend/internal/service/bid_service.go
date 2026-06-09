package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
)

type BidService struct {
	store       *memoryStore
	engine      *AuctionEngine
	runtime     *realtime.Runtime
	repo        repository.BidRepository
	roomRepo    repository.RoomRepository
	itemRepo    repository.ItemRepository
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

type CreateBidInput struct {
	RoomID    string
	SessionID string
	ItemID    string
	UserID    string
	BidPrice  int64
	RequestID string
}

type ConfigureAutoProxyInput struct {
	RoomID    string
	SessionID string
	ItemID    string
	UserID    string
	MaxPrice  int64
	Enabled   bool
}

func NewBidService(runtime *realtime.Runtime, repo repository.BidRepository, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, userRepo repository.UserRepository, sessionRepo repository.SessionRepository, resultRepo repository.ResultRepository, orderRepo repository.OrderRepository) *BidService {
	return &BidService{
		store:       sharedStore,
		engine:      NewAuctionEngine(sharedStore, runtime, nil, nil, sessionRepo, resultRepo, orderRepo),
		runtime:     runtime,
		repo:        repo,
		roomRepo:    roomRepo,
		itemRepo:    itemRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *BidService) CreateBid(input CreateBidInput) (model.BidResult, *SessionSettlement, error) {
	room, session, item, err := s.loadBidContext(input)
	if err != nil {
		return model.BidResult{}, nil, err
	}

	if room.CurrentSessionID != session.ID || session.RoomID != room.ID || session.ItemID != item.ID || item.RoomID != room.ID {
		return model.BidResult{}, nil, ErrBidOwnershipMismatch
	}

	if session.LeaderUserID != "" && session.LeaderUserID == input.UserID {
		return model.BidResult{}, nil, ErrAlreadyLeadingBid
	}

	if s.runtime == nil {
		return s.createBidWithMemoryLock(input, session, item)
	}

	var (
		acceptedBidPrice int64
		ceilingReached   bool
		extensionApplied bool
		nextMinimumBid   int64
	)

	runtimeResult, err := s.runtime.RunAtomicBid(realtime.AtomicBidInput{
		SessionID: input.SessionID,
		UserID:    input.UserID,
		BidPrice:  input.BidPrice,
		RequestID: input.RequestID,
		NowUnix:   time.Now().Unix(),
	})
	if err != nil {
		return model.BidResult{}, nil, err
	}
	if runtimeResult.Code == "duplicate_request" {
		if cached, exists := s.loadProcessedRequest(input.RequestID); exists {
			return cached, nil, ErrDuplicateBidRequest
		}
		return model.BidResult{}, nil, ErrDuplicateBidRequest
	}
	if !runtimeResult.OK {
		switch runtimeResult.Code {
		case "session_not_bidding":
			return model.BidResult{}, nil, ErrSessionNotBidding
		default:
			return model.BidResult{}, nil, ErrInvalidBidPrice
		}
	}

	acceptedBidPrice = runtimeResult.AcceptedBidPrice
	ceilingReached = runtimeResult.CeilingReached
	extensionApplied = runtimeResult.ExtensionApplied
	nextMinimumBid = runtimeResult.NextMinimumBid
	session.CurrentPrice = runtimeResult.CurrentPrice
	session.LeaderUserID = input.UserID
	session.ParticipantCount = runtimeResult.ParticipantCount
	if runtimeResult.EndTimeUnix > 0 {
		session.EndTime = time.Unix(runtimeResult.EndTimeUnix, 0).Format(time.RFC3339)
	}

	bid := model.Bid{
		ID:         nextBidID(s.loadBidCount()),
		SessionID:  input.SessionID,
		RoomID:     input.RoomID,
		ItemID:     input.ItemID,
		UserID:     input.UserID,
		BidPrice:   acceptedBidPrice,
		RequestID:  input.RequestID,
		RankAfter:  1,
		Status:     domain.BidStatusAccepted,
		CreateTime: time.Now().Format(time.RFC3339),
	}
	if s.runtime != nil {
		entry, ok, rankErr := s.runtime.GetUserRankingEntry(input.SessionID, input.UserID, s.loadUsers())
		if rankErr == nil && ok {
			bid.RankAfter = entry.Rank
		}
	}

	if s.repo != nil {
		if err := s.repo.CreateBid(bid); err != nil {
			return model.BidResult{}, nil, err
		}
	}
	if s.sessionRepo != nil {
		if err := s.sessionRepo.SaveSession(session); err != nil {
			return model.BidResult{}, nil, err
		}
	}

	var settlement *SessionSettlement
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
		NextMinimumBid:    nextMinimumBid,
		VibrateSignalHint: "overtake",
	}

	s.store.mu.Lock()
	if currentRoom, ok := s.store.rooms[session.RoomID]; ok && currentRoom.OnlineCount > session.ViewerCount {
		session.ViewerCount = currentRoom.OnlineCount
	}
	s.store.bids = append(s.store.bids, bid)
	s.store.sessions[session.ID] = session
	s.store.processedRequests[input.RequestID] = result
	if ceilingReached {
		outcome, err := s.engine.ReachCeilingLocked(session.ID)
		if err != nil {
			s.store.mu.Unlock()
			return model.BidResult{}, nil, err
		}
		settlement = &outcome
	}
	s.store.mu.Unlock()

	return result, settlement, nil
}

func (s *BidService) ConfigureAutoProxy(input ConfigureAutoProxyInput) (model.AutoProxyConfig, error) {
	_, session, item, err := s.loadBidContext(CreateBidInput{
		RoomID:    input.RoomID,
		SessionID: input.SessionID,
		ItemID:    input.ItemID,
		UserID:    input.UserID,
	})
	if err != nil {
		return model.AutoProxyConfig{}, err
	}

	key := autoProxyKey(input.SessionID, input.UserID)
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	if !input.Enabled {
		delete(s.store.autoProxyConfigs, key)
		return model.AutoProxyConfig{
			SessionID: input.SessionID,
			RoomID:    input.RoomID,
			ItemID:    input.ItemID,
			UserID:    input.UserID,
			MaxPrice:  0,
		}, nil
	}

	minAllowed := session.CurrentPrice + session.IncrementStep
	if session.CurrentPrice == 0 {
		minAllowed = item.StartPrice
	}
	if input.MaxPrice < minAllowed {
		return model.AutoProxyConfig{}, ErrInvalidBidPrice
	}
	if session.IncrementStep > 0 {
		delta := input.MaxPrice - session.CurrentPrice
		if delta > 0 && delta%session.IncrementStep != 0 {
			return model.AutoProxyConfig{}, ErrInvalidBidPrice
		}
	}

	now := time.Now().Format(time.RFC3339)
	config := model.AutoProxyConfig{
		SessionID: input.SessionID,
		RoomID:    input.RoomID,
		ItemID:    input.ItemID,
		UserID:    input.UserID,
		MaxPrice:  input.MaxPrice,
		EnabledAt: now,
	}
	if existing, ok := s.store.autoProxyConfigs[key]; ok && existing.EnabledAt != "" && existing.MaxPrice == input.MaxPrice {
		config.EnabledAt = existing.EnabledAt
	}
	s.store.autoProxyConfigs[key] = config
	return config, nil
}

func (s *BidService) GetAutoProxy(sessionID string, userID string) (model.AutoProxyConfig, bool) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	config, ok := s.store.autoProxyConfigs[autoProxyKey(sessionID, userID)]
	return config, ok
}

func (s *BidService) ProcessAutoProxy(roomID, sessionID, itemID, triggerUserID string) ([]model.BidResult, *SessionSettlement, error) {
	results := make([]model.BidResult, 0)
	for attempt := 0; attempt < 16; attempt++ {
		candidate, ok := s.pickAutoProxyCandidate(sessionID, triggerUserID)
		if !ok {
			return results, nil, nil
		}

		requestID := fmt.Sprintf("proxy-%d-%d", time.Now().UnixNano(), attempt)
		result, settlement, err := s.CreateBid(CreateBidInput{
			RoomID:    roomID,
			SessionID: sessionID,
			ItemID:    itemID,
			UserID:    candidate.userID,
			BidPrice:  candidate.bidPrice,
			RequestID: requestID,
		})
		if err != nil {
			// Skip stale candidates caused by concurrent updates and continue selecting.
			if errors.Is(err, ErrAlreadyLeadingBid) || errors.Is(err, ErrInvalidBidPrice) || errors.Is(err, ErrSessionNotBidding) {
				continue
			}
			return results, settlement, err
		}
		results = append(results, result)
		triggerUserID = candidate.userID
		if settlement != nil {
			return results, settlement, nil
		}
	}

	return results, nil, nil
}

func (s *BidService) createBidWithMemoryLock(input CreateBidInput, session model.AuctionSession, item model.AuctionItem) (model.BidResult, *SessionSettlement, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

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
		if err == nil && endTime.Sub(time.Now()) <= time.Duration(session.ExtensionTrigger)*time.Second {
			session.EndTime = endTime.Add(time.Duration(session.ExtensionSeconds) * time.Second).Format(time.RFC3339)
			extensionApplied = true
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

	rankings := buildRankings(append(append([]model.Bid(nil), s.store.bids...), bid), s.store.users, input.SessionID)
	for _, entry := range rankings {
		if entry.UserID == input.UserID {
			bid.RankAfter = entry.Rank
			break
		}
	}

	if s.repo != nil {
		if err := s.repo.CreateBid(bid); err != nil {
			return model.BidResult{}, nil, err
		}
	}
	if s.sessionRepo != nil {
		if err := s.sessionRepo.SaveSession(session); err != nil {
			return model.BidResult{}, nil, err
		}
	}

	s.store.bids = append(s.store.bids, bid)
	s.store.sessions[session.ID] = session

	var settlement *SessionSettlement
	if ceilingReached {
		outcome, err := s.engine.ReachCeilingLocked(session.ID)
		if err != nil {
			return model.BidResult{}, nil, err
		}
		settlement = &outcome
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

func (s *BidService) loadBidContext(input CreateBidInput) (model.LiveRoom, model.AuctionSession, model.AuctionItem, error) {
	if !s.userExists(input.UserID) {
		return model.LiveRoom{}, model.AuctionSession{}, model.AuctionItem{}, ErrUserNotFound
	}

	room, ok := s.loadRoom(input.RoomID)
	if !ok {
		return model.LiveRoom{}, model.AuctionSession{}, model.AuctionItem{}, ErrRoomNotFound
	}

	session, ok := s.loadSession(input.RoomID, input.SessionID)
	if !ok {
		return model.LiveRoom{}, model.AuctionSession{}, model.AuctionItem{}, ErrSessionNotFound
	}

	item, ok := s.loadItem(input.RoomID, input.ItemID)
	if !ok {
		return model.LiveRoom{}, model.AuctionSession{}, model.AuctionItem{}, ErrItemNotFound
	}

	return room, session, item, nil
}

func (s *BidService) userExists(userID string) bool {
	if userID == "" {
		return false
	}

	if s.userRepo != nil {
		user, err := s.userRepo.GetByID(userID)
		if err == nil && user != nil && user.ID != "" {
			return true
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	_, ok := s.store.users[userID]
	return ok
}

func (s *BidService) loadRoom(roomID string) (model.LiveRoom, bool) {
	if roomID == "" {
		return model.LiveRoom{}, false
	}

	if s.roomRepo != nil {
		room, err := s.roomRepo.GetRoomDetail(roomID)
		if err == nil && room != nil && room.ID != "" {
			return *room, true
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	room, ok := s.store.rooms[roomID]
	return room, ok
}

func (s *BidService) loadSession(roomID string, sessionID string) (model.AuctionSession, bool) {
	if sessionID == "" {
		return model.AuctionSession{}, false
	}

	if s.sessionRepo != nil && roomID != "" {
		sessions, err := s.sessionRepo.ListRoomSessions(roomID)
		if err == nil {
			for _, session := range sessions {
				if session.ID == sessionID {
					return session, true
				}
			}
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	session, ok := s.store.sessions[sessionID]
	return session, ok
}

func (s *BidService) loadItem(roomID string, itemID string) (model.AuctionItem, bool) {
	if itemID == "" {
		return model.AuctionItem{}, false
	}

	if s.itemRepo != nil && roomID != "" {
		item, err := s.itemRepo.GetItemDetail(roomID, itemID)
		if err == nil && item != nil && item.ID != "" {
			return *item, true
		}
	}

	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	item, ok := s.store.items[itemID]
	return item, ok
}

func (s *BidService) loadProcessedRequest(requestID string) (model.BidResult, bool) {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	result, ok := s.store.processedRequests[requestID]
	return result, ok
}

func (s *BidService) loadUsers() map[string]model.User {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	users := make(map[string]model.User, len(s.store.users))
	for id, user := range s.store.users {
		users[id] = user
	}
	return users
}

func (s *BidService) loadBidCount() int {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	return len(s.store.bids)
}

func (s *BidService) ListMyBids(userID string) []model.Bid {
	if s.repo != nil {
		if bids, err := s.repo.ListUserBids(userID); err == nil && len(bids) > 0 {
			return bids
		}
	}

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

type autoProxyCandidate struct {
	userID   string
	bidPrice int64
}

func (s *BidService) pickAutoProxyCandidate(sessionID string, triggerUserID string) (autoProxyCandidate, bool) {
	_ = triggerUserID

	s.store.mu.RLock()
	session, ok := s.store.sessions[sessionID]
	if !ok || session.Status != domain.SessionStateBidding || !session.SupportsAutoProxy {
		s.store.mu.RUnlock()
		return autoProxyCandidate{}, false
	}

	nextMinimumBid := session.CurrentPrice + session.IncrementStep
	incrementStep := session.IncrementStep
	type contender struct {
		config model.AutoProxyConfig
	}
	contenders := make([]contender, 0)
	for _, cfg := range s.store.autoProxyConfigs {
		if cfg.SessionID != sessionID || cfg.MaxPrice < nextMinimumBid {
			continue
		}
		if cfg.UserID == session.LeaderUserID {
			continue
		}
		contenders = append(contenders, contender{config: cfg})
	}
	s.store.mu.RUnlock()

	if len(contenders) == 0 {
		return autoProxyCandidate{}, false
	}

	sort.SliceStable(contenders, func(i, j int) bool {
		if contenders[i].config.MaxPrice != contenders[j].config.MaxPrice {
			return contenders[i].config.MaxPrice > contenders[j].config.MaxPrice
		}
		return contenders[i].config.EnabledAt < contenders[j].config.EnabledAt
	})

	top := contenders[0].config
	targetPrice := nextMinimumBid

	if len(contenders) > 1 {
		second := contenders[1].config
		if top.MaxPrice == second.MaxPrice {
			targetPrice = top.MaxPrice
		} else {
			step := incrementStep
			if step <= 0 {
				step = 1
			}
			targetPrice = second.MaxPrice + step
			if targetPrice < nextMinimumBid {
				targetPrice = nextMinimumBid
			}
			if targetPrice > top.MaxPrice {
				targetPrice = top.MaxPrice
			}
		}
	}

	if targetPrice < nextMinimumBid {
		targetPrice = nextMinimumBid
	}
	if targetPrice > top.MaxPrice {
		targetPrice = top.MaxPrice
	}
	if targetPrice <= 0 {
		return autoProxyCandidate{}, false
	}

	return autoProxyCandidate{
		userID:   top.UserID,
		bidPrice: targetPrice,
	}, true
}

func autoProxyKey(sessionID string, userID string) string {
	return sessionID + "::" + userID
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
