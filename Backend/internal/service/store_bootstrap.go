package service

import (
	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

func SyncMemoryBootstrap(store *memoryStore, userRepo repository.UserRepository, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, sessionRepo repository.SessionRepository, bidRepo repository.BidRepository) error {
	if store == nil || roomRepo == nil || itemRepo == nil || sessionRepo == nil {
		return nil
	}

	rooms, err := roomRepo.ListRooms()
	if err != nil || len(rooms) == 0 {
		return err
	}

	usersByID := make(map[string]model.User)
	userIDsByUsername := make(map[string]string)
	if userRepo != nil {
		users, err := userRepo.List()
		if err != nil {
			return err
		}
		for _, user := range users {
			usersByID[user.ID] = user
			if user.Username != "" {
				userIDsByUsername[user.Username] = user.ID
			}
		}
	}

	roomsByID := make(map[string]model.LiveRoom, len(rooms))
	itemsByID := make(map[string]model.AuctionItem)
	sessionsByID := make(map[string]model.AuctionSession)
	roomItems := make(map[string][]string, len(rooms))
	allBids := make([]model.Bid, 0)

	for _, room := range rooms {
		roomsByID[room.ID] = room

		items, err := itemRepo.ListRoomItems(room.ID)
		if err != nil {
			return err
		}
		itemIDs := make([]string, 0, len(items))
		for _, item := range items {
			itemsByID[item.ID] = item
			itemIDs = append(itemIDs, item.ID)
		}
		roomItems[room.ID] = itemIDs

		sessions, err := sessionRepo.ListRoomSessions(room.ID)
		if err != nil {
			return err
		}
		for _, session := range sessions {
			sessionsByID[session.ID] = session
			if bidRepo == nil {
				continue
			}
			bids, err := bidRepo.ListSessionBids(session.ID)
			if err != nil {
				return err
			}
			allBids = append(allBids, bids...)
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	if len(usersByID) > 0 {
		store.users = usersByID
		store.userIDsByUsername = userIDsByUsername
	}
	store.rooms = roomsByID
	store.items = itemsByID
	store.sessions = sessionsByID
	store.roomItems = roomItems
	store.bids = allBids
	store.orders = []model.AuctionOrder{}
	store.ordersByID = map[string]model.AuctionOrder{}
	store.ordersBySession = map[string]model.AuctionOrder{}
	store.processedRequests = map[string]model.BidResult{}

	// Rebuild online room status for persisted rooms that have an active session.
	for roomID, room := range store.rooms {
		if room.CurrentSessionID == "" {
			continue
		}
		session, ok := store.sessions[room.CurrentSessionID]
		if !ok {
			continue
		}
		if session.Status == domain.SessionStateBidding {
			room.Status = domain.RoomStatusLive
			store.rooms[roomID] = room
		}
	}

	return nil
}
