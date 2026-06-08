package service

import (
	"fmt"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
)

type AdminService struct {
	store       *memoryStore
	engine      *AuctionEngine
	runtime     *realtime.Runtime
	itemRepo    repository.ItemRepository
	sessionRepo repository.SessionRepository
}

func NewAdminService(runtime *realtime.Runtime, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, sessionRepo repository.SessionRepository, resultRepo repository.ResultRepository, orderRepo repository.OrderRepository) *AdminService {
	return &AdminService{
		store:       sharedStore,
		engine:      NewAuctionEngine(sharedStore, runtime, roomRepo, itemRepo, sessionRepo, resultRepo, orderRepo),
		runtime:     runtime,
		itemRepo:    itemRepo,
		sessionRepo: sessionRepo,
	}
}

type CreateItemInput struct {
	Title                   string
	CoverImage              string
	Description             string
	StartPrice              int64
	IncrementStep           int64
	CeilingPrice            *int64
	DurationSeconds         int
	ExtensionSeconds        int
	ExtensionTriggerSeconds int
}

type UpdateItemInput struct {
	Title                   *string
	CoverImage              *string
	Description             *string
	StartPrice              *int64
	IncrementStep           *int64
	CeilingPrice            *int64
	DurationSeconds         *int
	ExtensionSeconds        *int
	ExtensionTriggerSeconds *int
}

func (s *AdminService) CreateItem(roomID string, input CreateItemInput) (model.AuctionItem, map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	room, ok := s.store.rooms[roomID]
	if !ok || room.ID == "" {
		return model.AuctionItem{}, nil, ErrRoomNotFound
	}

	if input.Title == "" || input.IncrementStep <= 0 || input.DurationSeconds <= 0 {
		return model.AuctionItem{}, nil, ErrInvalidBidPrice
	}

	itemID := fmt.Sprintf("item-%03d", len(s.store.items)+1)
	sessionID := fmt.Sprintf("session-%03d", len(s.store.sessions)+1)
	previousQueue := append([]string(nil), s.store.roomItems[roomID]...)
	item := model.AuctionItem{
		ID:                      itemID,
		RoomID:                  roomID,
		Title:                   input.Title,
		CoverImage:              input.CoverImage,
		Description:             input.Description,
		StartPrice:              input.StartPrice,
		IncrementStep:           input.IncrementStep,
		CeilingPrice:            input.CeilingPrice,
		DurationSeconds:         input.DurationSeconds,
		ExtensionSeconds:        input.ExtensionSeconds,
		ExtensionTriggerSeconds: input.ExtensionTriggerSeconds,
		QueueStatus:             domain.QueueStateQueued,
	}
	s.store.items[itemID] = item
	s.store.roomItems[roomID] = append(s.store.roomItems[roomID], itemID)
	s.store.roomItems[roomID] = uniqueQueueIDs(s.store.roomItems[roomID])
	session := model.AuctionSession{
		ID:                sessionID,
		RoomID:            roomID,
		ItemID:            itemID,
		Status:            domain.SessionStatePending,
		CurrentPrice:      input.StartPrice,
		LeaderUserID:      "",
		EndTime:           "",
		ParticipantCount:  0,
		IncrementStep:     input.IncrementStep,
		ExtensionSeconds:  input.ExtensionSeconds,
		ExtensionTrigger:  input.ExtensionTriggerSeconds,
		CeilingPrice:      input.CeilingPrice,
		SupportsAutoProxy: true,
	}
	s.store.sessions[sessionID] = session
	if s.itemRepo != nil {
		if err := s.itemRepo.SaveItem(item); err != nil {
			delete(s.store.items, itemID)
			delete(s.store.sessions, sessionID)
			s.store.roomItems[roomID] = previousQueue
			return model.AuctionItem{}, nil, err
		}
		if err := s.itemRepo.ReplaceRoomQueue(roomID, s.store.roomItems[roomID]); err != nil {
			delete(s.store.items, itemID)
			delete(s.store.sessions, sessionID)
			s.store.roomItems[roomID] = previousQueue
			return model.AuctionItem{}, nil, err
		}
	}
	if s.sessionRepo != nil {
		if err := s.sessionRepo.SaveSession(session); err != nil {
			delete(s.store.items, itemID)
			delete(s.store.sessions, sessionID)
			s.store.roomItems[roomID] = previousQueue
			return model.AuctionItem{}, nil, err
		}
	}
	if s.runtime != nil {
		if err := saveSessionState(s.runtime, session, item); err != nil {
			delete(s.store.items, itemID)
			delete(s.store.sessions, sessionID)
			s.store.roomItems[roomID] = previousQueue
			return model.AuctionItem{}, nil, err
		}
	}

	meta := map[string]any{
		"roomId":    roomID,
		"itemId":    itemID,
		"sessionId": sessionID,
		"queueSize": len(s.store.roomItems[roomID]),
	}

	return item, meta, nil
}

func (s *AdminService) UpdateItem(itemID string, input UpdateItemInput) (model.AuctionItem, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	item, ok := s.store.items[itemID]
	if !ok {
		return model.AuctionItem{}, ErrItemNotFound
	}
	if item.QueueStatus == domain.QueueStateActive || item.QueueStatus == domain.QueueStateFinished || item.QueueStatus == domain.QueueStateCancelled {
		return model.AuctionItem{}, ErrInvalidSessionState
	}

	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.CoverImage != nil {
		item.CoverImage = *input.CoverImage
	}
	if input.Description != nil {
		item.Description = *input.Description
	}
	if input.StartPrice != nil {
		item.StartPrice = *input.StartPrice
	}
	if input.IncrementStep != nil && *input.IncrementStep > 0 {
		item.IncrementStep = *input.IncrementStep
	}
	if input.CeilingPrice != nil {
		item.CeilingPrice = input.CeilingPrice
	}
	if input.DurationSeconds != nil && *input.DurationSeconds > 0 {
		item.DurationSeconds = *input.DurationSeconds
	}
	if input.ExtensionSeconds != nil && *input.ExtensionSeconds >= 0 {
		item.ExtensionSeconds = *input.ExtensionSeconds
	}
	if input.ExtensionTriggerSeconds != nil && *input.ExtensionTriggerSeconds >= 0 {
		item.ExtensionTriggerSeconds = *input.ExtensionTriggerSeconds
	}

	s.store.items[itemID] = item
	if s.itemRepo != nil {
		if err := s.itemRepo.SaveItem(item); err != nil {
			return model.AuctionItem{}, err
		}
	}
	for sessionID, session := range s.store.sessions {
		if session.ItemID == itemID && session.Status == domain.SessionStatePending {
			session.CurrentPrice = item.StartPrice
			session.IncrementStep = item.IncrementStep
			session.ExtensionSeconds = item.ExtensionSeconds
			session.ExtensionTrigger = item.ExtensionTriggerSeconds
			session.CeilingPrice = item.CeilingPrice
			s.store.sessions[sessionID] = session
			if s.sessionRepo != nil {
				if err := s.sessionRepo.SaveSession(session); err != nil {
					return model.AuctionItem{}, err
				}
			}
			if s.runtime != nil {
				if err := saveSessionState(s.runtime, session, item); err != nil {
					return model.AuctionItem{}, err
				}
			}
		}
	}

	return item, nil
}

func (s *AdminService) ReorderQueue(roomID string, itemIDs []string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	current := s.store.roomItems[roomID]
	if len(current) == 0 {
		return nil, ErrRoomNotFound
	}

	if len(itemIDs) == 0 {
		return map[string]any{
			"roomId": roomID,
			"items":  current,
		}, nil
	}

	if len(itemIDs) != len(current) {
		return nil, ErrInvalidQueueOrder
	}

	allowed := map[string]struct{}{}
	for _, itemID := range current {
		allowed[itemID] = struct{}{}
	}
	for _, itemID := range itemIDs {
		if _, ok := allowed[itemID]; !ok {
			return nil, ErrInvalidQueueOrder
		}
	}

	activeCount := 0
	upcomingCount := 0
	for _, itemID := range itemIDs {
		status := s.store.items[itemID].QueueStatus
		if status == domain.QueueStateFinished || status == domain.QueueStateCancelled {
			return nil, ErrInvalidQueueOrder
		}
		if status == domain.QueueStateActive {
			activeCount++
		}
		if status == domain.QueueStateUpcoming {
			upcomingCount++
		}
	}
	if activeCount > 1 || upcomingCount > 1 {
		return nil, ErrInvalidQueueOrder
	}

	s.store.roomItems[roomID] = itemIDs
	if s.itemRepo != nil {
		if err := s.itemRepo.ReplaceRoomQueue(roomID, itemIDs); err != nil {
			return nil, err
		}
	}
	return map[string]any{
		"roomId": roomID,
		"items":  itemIDs,
	}, nil
}

func (s *AdminService) ActivateNextItem(roomID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	return s.engine.ActivateNextItemLocked(roomID)
}

func (s *AdminService) StartSession(sessionID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	return s.engine.StartSessionLocked(sessionID)
}

func (s *AdminService) StartRoomLive(roomID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	room, ok := s.store.rooms[roomID]
	if !ok || room.ID == "" {
		return nil, ErrRoomNotFound
	}

	room.Status = domain.RoomStatusLive
	s.store.rooms[roomID] = room
	if s.engine != nil && s.engine.roomRepo != nil {
		if err := s.engine.roomRepo.SaveRoom(room); err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"roomId": roomID,
		"status": room.Status,
	}, nil
}

func (s *AdminService) StopRoomLive(roomID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	room, ok := s.store.rooms[roomID]
	if !ok || room.ID == "" {
		return nil, ErrRoomNotFound
	}

	for _, session := range s.store.sessions {
		if session.RoomID == roomID && session.Status == domain.SessionStateBidding {
			return nil, ErrRoomHasActiveSession
		}
	}

	room.Status = domain.RoomStatusOffline
	s.store.rooms[roomID] = room
	if s.engine != nil && s.engine.roomRepo != nil {
		if err := s.engine.roomRepo.SaveRoom(room); err != nil {
			return nil, err
		}
	}

	return map[string]any{
		"roomId": roomID,
		"status": room.Status,
	}, nil
}

func (s *AdminService) CancelSession(sessionID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	s.refreshSessionContextLocked(sessionID)
	return s.engine.CancelSessionLocked(sessionID)
}

func (s *AdminService) SettleSession(sessionID string) (SessionSettlement, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	s.refreshSessionContextLocked(sessionID)
	return s.engine.SettleSessionLocked(sessionID)
}

func (s *AdminService) refreshSessionContextLocked(sessionID string) {
	if sessionID == "" || s.sessionRepo == nil {
		return
	}

	for _, roomID := range s.store.roomItems {
		_ = roomID
	}

	for roomID := range s.store.rooms {
		sessions, err := s.sessionRepo.ListRoomSessions(roomID)
		if err != nil {
			continue
		}
		for _, session := range sessions {
			if session.ID != sessionID {
				continue
			}
			s.store.sessions[session.ID] = session
			if s.itemRepo != nil {
				if item, err := s.itemRepo.GetItemDetail(session.RoomID, session.ItemID); err == nil && item != nil && item.ID != "" {
					s.store.items[item.ID] = *item
				}
			}
			if room, ok := s.store.rooms[session.RoomID]; ok {
				if room.CurrentSessionID == "" {
					room.CurrentSessionID = session.ID
					s.store.rooms[session.RoomID] = room
				}
			}
			return
		}
	}
}

func (s *AdminService) ListRoomSessions(roomID string) []map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	results := make([]map[string]any, 0)
	for _, itemID := range s.store.roomItems[roomID] {
		for sessionID, session := range s.store.sessions {
			if session.RoomID == roomID && session.ItemID == itemID {
				results = append(results, map[string]any{
					"roomId":       roomID,
					"sessionId":    sessionID,
					"itemId":       session.ItemID,
					"status":       session.Status,
					"currentPrice": session.CurrentPrice,
					"queueStatus":  s.store.items[itemID].QueueStatus,
					"endTime":      session.EndTime,
				})
			}
		}
	}
	return results
}

func (s *AdminService) ListOrders() []map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	orders := make([]map[string]any, 0, len(s.store.orders))
	for _, order := range s.store.orders {
		orders = append(orders, map[string]any{
			"orderId":     order.ID,
			"sessionId":   order.SessionID,
			"roomId":      order.RoomID,
			"itemId":      order.ItemID,
			"buyerUserId": order.BuyerUserID,
			"amount":      order.Amount,
			"status":      order.Status,
			"createTime":  order.CreateTime,
		})
	}
	return orders
}

func (s *AdminService) ListOrdersByAnchorUserID(anchorUserID string) []map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	orders := make([]map[string]any, 0, len(s.store.orders))
	for _, order := range s.store.orders {
		room, ok := s.store.rooms[order.RoomID]
		if !ok || room.AnchorUserID != anchorUserID {
			continue
		}
		orders = append(orders, map[string]any{
			"orderId":     order.ID,
			"sessionId":   order.SessionID,
			"roomId":      order.RoomID,
			"itemId":      order.ItemID,
			"buyerUserId": order.BuyerUserID,
			"amount":      order.Amount,
			"status":      order.Status,
			"createTime":  order.CreateTime,
		})
	}
	return orders
}

func (s *AdminService) GetStatsOverview() map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	sold := 0
	cancelled := 0
	for _, session := range s.store.sessions {
		if session.Status == domain.SessionStateEndedSold {
			sold++
		}
		if session.Status == domain.SessionStateCancelled {
			cancelled++
		}
	}

	return map[string]any{
		"totalRooms":        len(s.store.rooms),
		"totalSessions":     len(s.store.sessions),
		"soldSessions":      sold,
		"cancelledSessions": cancelled,
	}
}

func (s *AdminService) GetStatsOverviewByAnchorUserID(anchorUserID string) map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	roomCount := 0
	sessionCount := 0
	sold := 0
	cancelled := 0
	for _, room := range s.store.rooms {
		if room.AnchorUserID != anchorUserID {
			continue
		}
		roomCount++
		for _, session := range s.store.sessions {
			if session.RoomID != room.ID {
				continue
			}
			sessionCount++
			if session.Status == domain.SessionStateEndedSold {
				sold++
			}
			if session.Status == domain.SessionStateCancelled {
				cancelled++
			}
		}
	}

	return map[string]any{
		"totalRooms":        roomCount,
		"totalSessions":     sessionCount,
		"soldSessions":      sold,
		"cancelledSessions": cancelled,
	}
}

func (s *AdminService) GetStatsTimeline() []map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	timeline := make([]map[string]any, 0, len(s.store.bids))
	for _, bid := range s.store.bids {
		timeline = append(timeline, map[string]any{
			"time":      bid.CreateTime,
			"event":     "bid_accepted",
			"sessionId": bid.SessionID,
			"itemId":    bid.ItemID,
			"price":     bid.BidPrice,
			"userId":    bid.UserID,
		})
	}
	return timeline
}

func (s *AdminService) GetStatsTimelineByAnchorUserID(anchorUserID string) []map[string]any {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	timeline := make([]map[string]any, 0, len(s.store.bids))
	for _, bid := range s.store.bids {
		room, ok := s.store.rooms[bid.RoomID]
		if !ok || room.AnchorUserID != anchorUserID {
			continue
		}
		timeline = append(timeline, map[string]any{
			"time":      bid.CreateTime,
			"event":     "bid_accepted",
			"sessionId": bid.SessionID,
			"itemId":    bid.ItemID,
			"price":     bid.BidPrice,
			"userId":    bid.UserID,
		})
	}
	return timeline
}

func (s *AdminService) RoomOwnedBy(roomID string, anchorUserID string) bool {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	room, ok := s.store.rooms[roomID]
	return ok && room.AnchorUserID == anchorUserID
}

func (s *AdminService) SessionOwnedBy(sessionID string, anchorUserID string) bool {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	session, ok := s.store.sessions[sessionID]
	if !ok {
		return false
	}
	room, ok := s.store.rooms[session.RoomID]
	return ok && room.AnchorUserID == anchorUserID
}

func (s *AdminService) ItemOwnedBy(itemID string, anchorUserID string) bool {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	item, ok := s.store.items[itemID]
	if !ok {
		return false
	}
	room, ok := s.store.rooms[item.RoomID]
	return ok && room.AnchorUserID == anchorUserID
}

func (s *AdminService) OrderOwnedBy(orderID string, anchorUserID string) bool {
	s.store.mu.RLock()
	defer s.store.mu.RUnlock()

	order, ok := s.store.ordersByID[orderID]
	if !ok {
		return false
	}
	room, ok := s.store.rooms[order.RoomID]
	return ok && room.AnchorUserID == anchorUserID
}

func uniqueQueueIDs(itemIDs []string) []string {
	if len(itemIDs) <= 1 {
		return itemIDs
	}

	seen := make(map[string]struct{}, len(itemIDs))
	result := make([]string, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		if itemID == "" {
			continue
		}
		if _, exists := seen[itemID]; exists {
			continue
		}
		seen[itemID] = struct{}{}
		result = append(result, itemID)
	}
	return result
}
