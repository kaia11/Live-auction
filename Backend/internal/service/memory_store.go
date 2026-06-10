package service

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
)

type memoryStore struct {
	mu                sync.RWMutex
	users             map[string]model.User
	userIDsByUsername map[string]string
	rooms             map[string]model.LiveRoom
	items             map[string]model.AuctionItem
	sessions          map[string]model.AuctionSession
	roomItems         map[string][]string
	bids              []model.Bid
	orders            []model.AuctionOrder
	ordersByID        map[string]model.AuctionOrder
	ordersBySession   map[string]model.AuctionOrder
	processedRequests map[string]model.BidResult
	autoProxyConfigs  map[string]model.AutoProxyConfig
}

var sharedStore = newMemoryStore()
var bidIDCounter uint64

func SharedStore() *memoryStore {
	return sharedStore
}

func newMemoryStore() *memoryStore {
	ceilingPrice := int64(999)
	now := time.Now()

	store := &memoryStore{
		users: map[string]model.User{
			"user-001":   {ID: "user-001", Username: "silence", Password: "111111", Nickname: "阿宁", Avatar: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&q=80", Role: domain.UserRoleViewer},
			"user-002":   {ID: "user-002", Username: "viewer_guest", Password: "111111", Nickname: "小满", Avatar: "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200&q=80", Role: domain.UserRoleViewer},
			"user-003":   {ID: "user-003", Username: "viewer_vip", Password: "111111", Nickname: "阿青", Avatar: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=200&q=80", Role: domain.UserRoleViewer},
			"anchor-001": {ID: "anchor-001", Username: "silence001", Password: "111111", Nickname: "主播小玉", Avatar: "https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=200&q=80", Role: domain.UserRoleAnchor},
			"admin-001":  {ID: "admin-001", Username: "admin_root", Password: "111111", Nickname: "运营管理员", Avatar: "https://images.unsplash.com/photo-1519085360753-af0119f7cbe7?w=200&q=80", Role: domain.UserRoleAdmin},
		},
		userIDsByUsername: map[string]string{
			"silence":      "user-001",
			"viewer_guest": "user-002",
			"viewer_vip":   "user-003",
			"silence001":   "anchor-001",
			"admin_root":   "admin-001",
		},
		rooms: map[string]model.LiveRoom{
			"room-001": {
				ID:               "room-001",
				Title:            "古风首饰直播竞拍间",
				CoverImage:       "https://images.unsplash.com/photo-1617038220319-276d3cfab638?w=1200&q=80",
				VideoURL:         "https://www.w3schools.com/html/mov_bbb.mp4",
				Status:           domain.RoomStatusLive,
				AnchorUserID:     "anchor-001",
				AnchorName:       "主播小玉",
				OnlineCount:      1288,
				Thumbnail:        "https://images.unsplash.com/photo-1512436991641-6745cdb1723f?w=200&q=80",
				CurrentSessionID: "session-001",
			},
		},
		items: map[string]model.AuctionItem{
			"item-001": {
				ID:                      "item-001",
				RoomID:                  "room-001",
				Title:                   "和田玉吊坠",
				CoverImage:              "https://images.unsplash.com/photo-1512436991641-6745cdb1723f?w=800&q=80",
				Description:             "直播竞拍样例拍品，当前使用假视频占位进行联调。",
				StartPrice:              0,
				IncrementStep:           5,
				CeilingPrice:            &ceilingPrice,
				DurationSeconds:         120,
				ExtensionSeconds:        30,
				ExtensionTriggerSeconds: 30,
				QueueStatus:             domain.QueueStateActive,
			},
			"item-002": {
				ID:                      "item-002",
				RoomID:                  "room-001",
				Title:                   "鎏金花丝耳坠",
				CoverImage:              "https://images.unsplash.com/photo-1617038220319-276d3cfab638?w=800&q=80",
				Description:             "待上场拍品，用于后续拍品队列与切场开发。",
				StartPrice:              0,
				IncrementStep:           10,
				DurationSeconds:         120,
				ExtensionSeconds:        30,
				ExtensionTriggerSeconds: 30,
				QueueStatus:             domain.QueueStateQueued,
			},
		},
		sessions: map[string]model.AuctionSession{
			"session-001": {
				ID:                "session-001",
				RoomID:            "room-001",
				ItemID:            "item-001",
				Status:            domain.SessionStateBidding,
				CurrentPrice:      135,
				LeaderUserID:      "user-003",
				EndTime:           now.Add(2 * time.Minute).Format(time.RFC3339),
				ParticipantCount:  3,
				IncrementStep:     5,
				ExtensionSeconds:  30,
				ExtensionTrigger:  30,
				CeilingPrice:      &ceilingPrice,
				SupportsAutoProxy: true,
			},
			"session-002": {
				ID:                "session-002",
				RoomID:            "room-001",
				ItemID:            "item-002",
				Status:            domain.SessionStatePending,
				CurrentPrice:      0,
				LeaderUserID:      "",
				EndTime:           "",
				ParticipantCount:  0,
				IncrementStep:     10,
				ExtensionSeconds:  30,
				ExtensionTrigger:  30,
				CeilingPrice:      nil,
				SupportsAutoProxy: true,
			},
		},
		roomItems: map[string][]string{
			"room-001": {"item-001", "item-002"},
		},
		bids: []model.Bid{
			{ID: "bid-001", SessionID: "session-001", RoomID: "room-001", ItemID: "item-001", UserID: "user-001", BidPrice: 125, RequestID: "req-001", RankAfter: 3, Status: domain.BidStatusAccepted, CreateTime: now.Add(-90 * time.Second).Format(time.RFC3339)},
			{ID: "bid-002", SessionID: "session-001", RoomID: "room-001", ItemID: "item-001", UserID: "user-002", BidPrice: 130, RequestID: "req-002", RankAfter: 2, Status: domain.BidStatusAccepted, CreateTime: now.Add(-60 * time.Second).Format(time.RFC3339)},
			{ID: "bid-003", SessionID: "session-001", RoomID: "room-001", ItemID: "item-001", UserID: "user-003", BidPrice: 135, RequestID: "req-003", RankAfter: 1, Status: domain.BidStatusAccepted, CreateTime: now.Add(-30 * time.Second).Format(time.RFC3339)},
		},
		ordersByID:        map[string]model.AuctionOrder{},
		ordersBySession:   map[string]model.AuctionOrder{},
		processedRequests: map[string]model.BidResult{},
		autoProxyConfigs:  map[string]model.AutoProxyConfig{},
	}

	store.processedRequests["req-001"] = model.BidResult{
		RoomID:            "room-001",
		SessionID:         "session-001",
		ItemID:            "item-001",
		UserID:            "user-001",
		AcceptedBidPrice:  125,
		RequestID:         "req-001",
		CurrentPrice:      125,
		IsLeading:         false,
		ExtensionApplied:  false,
		CeilingReached:    false,
		NextMinimumBid:    130,
		VibrateSignalHint: "none",
	}

	return store
}

func nextBidID(_ int) string {
	return fmt.Sprintf("bid-%d", atomic.AddUint64(&bidIDCounter, 1))
}

func (s *memoryStore) ListRooms() []model.LiveRoom {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]model.LiveRoom, 0, len(s.rooms))
	for _, room := range s.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

func (s *memoryStore) ListRoomsByAnchorUserID(anchorUserID string) []model.LiveRoom {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rooms := make([]model.LiveRoom, 0)
	for _, room := range s.rooms {
		if room.AnchorUserID == anchorUserID {
			rooms = append(rooms, room)
		}
	}
	return rooms
}

func (s *memoryStore) GetRoomDetail(roomID string) model.LiveRoom {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.rooms[roomID]
}

func (s *memoryStore) ListRoomItems(roomID string) []model.AuctionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	itemIDs := s.roomItems[roomID]
	items := make([]model.AuctionItem, 0, len(itemIDs))
	for _, itemID := range itemIDs {
		items = append(items, s.items[itemID])
	}
	return items
}

func (s *memoryStore) GetItemDetail(roomID string, itemID string) model.AuctionItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item := s.items[itemID]
	if item.RoomID != roomID {
		return model.AuctionItem{}
	}
	return item
}

func (s *memoryStore) GetCurrentSession(roomID string) model.AuctionSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room := s.rooms[roomID]
	return s.sessions[room.CurrentSessionID]
}

func (s *memoryStore) ListRoomSessions(roomID string) []model.AuctionSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]model.AuctionSession, 0)
	for _, session := range s.sessions {
		if session.RoomID == roomID {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

func (s *memoryStore) ListUserOrders(userID string) []model.AuctionOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]model.AuctionOrder, 0)
	for _, order := range s.orders {
		if order.BuyerUserID == userID {
			orders = append(orders, order)
		}
	}
	return orders
}

func (s *memoryStore) ListUserBids(userID string) []model.Bid {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bids := make([]model.Bid, 0)
	for _, bid := range s.bids {
		if bid.UserID == userID {
			bids = append(bids, bid)
		}
	}
	return bids
}

func (s *memoryStore) ListSessionBids(sessionID string) []model.Bid {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bids := make([]model.Bid, 0)
	for _, bid := range s.bids {
		if bid.SessionID == sessionID {
			bids = append(bids, bid)
		}
	}
	return bids
}

func (s *memoryStore) ListUsers() []model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]model.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

func (s *memoryStore) GetUserByID(userID string) model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.users[userID]
}

func (s *memoryStore) GetUserByUsername(username string) model.User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	userID := s.userIDsByUsername[username]
	return s.users[userID]
}

func (s *memoryStore) SaveUser(user model.User) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.users[user.ID] = user
	s.userIDsByUsername[user.Username] = user.ID
}
