package service

import (
	"time"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
)

func SyncRealtimeBootstrap(runtime *realtime.Runtime, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, sessionRepo repository.SessionRepository, bidRepo repository.BidRepository, userRepo repository.UserRepository, fallback *memoryStore) error {
	if runtime == nil {
		return nil
	}

	if roomRepo != nil && itemRepo != nil && sessionRepo != nil {
		if ok, err := syncRealtimeFromRepositories(runtime, roomRepo, itemRepo, sessionRepo, bidRepo, userRepo, fallback); ok || err != nil {
			return err
		}
	}

	return syncRealtimeFromMemory(runtime, fallback)
}

func syncRealtimeFromRepositories(runtime *realtime.Runtime, roomRepo repository.RoomRepository, itemRepo repository.ItemRepository, sessionRepo repository.SessionRepository, bidRepo repository.BidRepository, userRepo repository.UserRepository, fallback *memoryStore) (bool, error) {
	rooms, err := roomRepo.ListRooms()
	if err != nil || len(rooms) == 0 {
		return false, err
	}

	userMap := copyUsers(fallback)
	for _, room := range rooms {
		if err := runtime.SetRoomCurrentSession(room.ID, room.CurrentSessionID); err != nil {
			return true, err
		}

		items, err := itemRepo.ListRoomItems(room.ID)
		if err != nil {
			return true, err
		}
		itemMap := make(map[string]model.AuctionItem, len(items))
		for _, item := range items {
			itemMap[item.ID] = item
		}

		sessions, err := sessionRepo.ListRoomSessions(room.ID)
		if err != nil {
			return true, err
		}
		for _, session := range sessions {
			item, ok := itemMap[session.ItemID]
			if !ok {
				if fallbackItem, found := fallback.items[session.ItemID]; found {
					item = fallbackItem
				} else {
					continue
				}
			}

			if err := saveSessionState(runtime, session, item); err != nil {
				return true, err
			}

			if bidRepo == nil {
				continue
			}
			bids, err := bidRepo.ListSessionBids(session.ID)
			if err != nil {
				return true, err
			}
			hydrateUsers(userMap, userRepo, bids)
			if err := runtime.ReplaceRanking(session.ID, buildRankings(bids, userMap, session.ID)); err != nil {
				return true, err
			}
		}
	}

	return true, nil
}

func syncRealtimeFromMemory(runtime *realtime.Runtime, store *memoryStore) error {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for roomID, room := range store.rooms {
		if err := runtime.SetRoomCurrentSession(roomID, room.CurrentSessionID); err != nil {
			return err
		}
	}

	for sessionID, session := range store.sessions {
		item := store.items[session.ItemID]
		if err := saveSessionState(runtime, session, item); err != nil {
			return err
		}
		rankings := buildRankings(store.bids, store.users, sessionID)
		if err := runtime.ReplaceRanking(sessionID, rankings); err != nil {
			return err
		}
	}

	return nil
}

func saveSessionState(runtime *realtime.Runtime, session model.AuctionSession, item model.AuctionItem) error {
	if runtime == nil {
		return nil
	}

	var endTimeUnix int64
	if session.EndTime != "" {
		if parsed, err := time.Parse(time.RFC3339, session.EndTime); err == nil {
			endTimeUnix = parsed.Unix()
		}
	}

	return runtime.SaveSessionState(realtime.SessionState{
		SessionID:               session.ID,
		RoomID:                  session.RoomID,
		ItemID:                  session.ItemID,
		Status:                  session.Status,
		CurrentPrice:            session.CurrentPrice,
		LeaderUserID:            session.LeaderUserID,
		ParticipantCount:        session.ParticipantCount,
		StartPrice:              item.StartPrice,
		IncrementStep:           session.IncrementStep,
		ExtensionSeconds:        session.ExtensionSeconds,
		ExtensionTriggerSeconds: session.ExtensionTrigger,
		CeilingPrice:            session.CeilingPrice,
		EndTimeUnix:             endTimeUnix,
	})
}

func syncRanking(runtime *realtime.Runtime, store *memoryStore, sessionID string) error {
	if runtime == nil {
		return nil
	}

	rankings := buildRankings(store.bids, store.users, sessionID)
	return runtime.ReplaceRanking(sessionID, rankings)
}

func copyUsers(store *memoryStore) map[string]model.User {
	store.mu.RLock()
	defer store.mu.RUnlock()

	users := make(map[string]model.User, len(store.users))
	for id, user := range store.users {
		users[id] = user
	}
	return users
}

func hydrateUsers(users map[string]model.User, repo repository.UserRepository, bids []model.Bid) {
	if repo == nil {
		return
	}

	for _, bid := range bids {
		if _, ok := users[bid.UserID]; ok {
			continue
		}
		user, err := repo.GetByID(bid.UserID)
		if err != nil || user == nil || user.ID == "" {
			continue
		}
		users[user.ID] = *user
	}
}
