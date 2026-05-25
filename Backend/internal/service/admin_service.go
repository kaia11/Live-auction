package service

import (
	"fmt"
	"time"

	"auction-live/backend/internal/model"
)

type AdminService struct {
	store *memoryStore
}

func NewAdminService() *AdminService {
	return &AdminService{store: sharedStore}
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
		QueueStatus:             "queued",
	}
	s.store.items[itemID] = item
	s.store.roomItems[roomID] = append(s.store.roomItems[roomID], itemID)
	s.store.sessions[sessionID] = model.AuctionSession{
		ID:                sessionID,
		RoomID:            roomID,
		ItemID:            itemID,
		Status:            "pending",
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
	if item.QueueStatus == "active" || item.QueueStatus == "finished" || item.QueueStatus == "cancelled" {
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
	for sessionID, session := range s.store.sessions {
		if session.ItemID == itemID && session.Status == "pending" {
			session.CurrentPrice = item.StartPrice
			session.IncrementStep = item.IncrementStep
			session.ExtensionSeconds = item.ExtensionSeconds
			session.ExtensionTrigger = item.ExtensionTriggerSeconds
			session.CeilingPrice = item.CeilingPrice
			s.store.sessions[sessionID] = session
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
		if status == "finished" || status == "cancelled" {
			return nil, ErrInvalidQueueOrder
		}
		if status == "active" {
			activeCount++
		}
		if status == "upcoming" {
			upcomingCount++
		}
	}
	if activeCount > 1 || upcomingCount > 1 {
		return nil, ErrInvalidQueueOrder
	}

	s.store.roomItems[roomID] = itemIDs
	return map[string]any{
		"roomId": roomID,
		"items":  itemIDs,
	}, nil
}

func (s *AdminService) ActivateNextItem(roomID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	room, ok := s.store.rooms[roomID]
	if !ok {
		return nil, ErrRoomNotFound
	}

	currentSession, ok := s.store.sessions[room.CurrentSessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	currentItem := s.store.items[currentSession.ItemID]
	if currentItem.QueueStatus == "active" {
		currentItem.QueueStatus = "cancelled"
		s.store.items[currentItem.ID] = currentItem
	}
	if currentSession.Status == "bidding" || currentSession.Status == "pending" {
		currentSession.Status = "cancelled"
		s.store.sessions[currentSession.ID] = currentSession
	}

	nextSessionID, nextItemID := s.findNextQueuedLocked(roomID)
	if nextSessionID == "" {
		return nil, ErrQueueExhausted
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

	room.CurrentSessionID = nextSessionID
	s.store.rooms[roomID] = room

	return map[string]any{
		"roomId":        roomID,
		"nextItemId":    nextItemID,
		"nextSessionId": nextSessionID,
		"status":        nextSession.Status,
		"queueStatus":   nextItem.QueueStatus,
	}, nil
}

func (s *AdminService) StartSession(sessionID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	session, ok := s.store.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if session.Status != "pending" {
		return nil, ErrInvalidSessionState
	}

	item, ok := s.store.items[session.ItemID]
	if !ok {
		return nil, ErrItemNotFound
	}

	session.Status = "bidding"
	session.CurrentPrice = item.StartPrice
	session.EndTime = time.Now().Add(time.Duration(item.DurationSeconds) * time.Second).Format(time.RFC3339)
	session.IncrementStep = item.IncrementStep
	session.ExtensionSeconds = item.ExtensionSeconds
	session.ExtensionTrigger = item.ExtensionTriggerSeconds
	session.CeilingPrice = item.CeilingPrice
	s.store.sessions[sessionID] = session

	item.QueueStatus = "active"
	s.store.items[item.ID] = item

	room := s.store.rooms[session.RoomID]
	room.CurrentSessionID = sessionID
	s.store.rooms[session.RoomID] = room

	return map[string]any{
		"roomId":      session.RoomID,
		"sessionId":   sessionID,
		"status":      session.Status,
		"endTime":     session.EndTime,
		"queueStatus": item.QueueStatus,
	}, nil
}

func (s *AdminService) CancelSession(sessionID string) (map[string]any, error) {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	session, ok := s.store.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if session.Status == "ended_sold" || session.Status == "ended_unsold" {
		return nil, ErrInvalidSessionState
	}

	session.Status = "cancelled"
	s.store.sessions[sessionID] = session

	item := s.store.items[session.ItemID]
	item.QueueStatus = "cancelled"
	s.store.items[item.ID] = item

	return map[string]any{
		"roomId":    session.RoomID,
		"sessionId": sessionID,
		"status":    session.Status,
	}, nil
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

	orders := make([]map[string]any, 0)
	for _, session := range s.store.sessions {
		if session.Status != "ended_sold" {
			continue
		}
		orders = append(orders, map[string]any{
			"orderId":     fmt.Sprintf("order-%s", session.ID),
			"itemId":      session.ItemID,
			"buyerUserId": session.LeaderUserID,
			"amount":      session.CurrentPrice,
			"status":      "pending_payment",
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
		if session.Status == "ended_sold" {
			sold++
		}
		if session.Status == "cancelled" {
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

func (s *AdminService) findNextQueuedLocked(roomID string) (string, string) {
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
