package service

import (
	"fmt"
	"sync"
	"time"

	"auction-live/backend/internal/model"
)

type memoryStore struct {
	mu                sync.RWMutex
	users             map[string]model.User
	rooms             map[string]model.LiveRoom
	items             map[string]model.AuctionItem
	sessions          map[string]model.AuctionSession
	roomItems         map[string][]string
	bids              []model.Bid
	processedRequests map[string]model.BidResult
}

var sharedStore = newMemoryStore()

func newMemoryStore() *memoryStore {
	ceilingPrice := int64(999)
	now := time.Now()

	store := &memoryStore{
		users: map[string]model.User{
			"user-001":   {ID: "user-001", Nickname: "阿宁", Avatar: "https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=200&q=80", Role: "viewer"},
			"user-002":   {ID: "user-002", Nickname: "小满", Avatar: "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=200&q=80", Role: "viewer"},
			"user-003":   {ID: "user-003", Nickname: "阿青", Avatar: "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=200&q=80", Role: "viewer"},
			"anchor-001": {ID: "anchor-001", Nickname: "主播小玉", Avatar: "https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=200&q=80", Role: "anchor"},
		},
		rooms: map[string]model.LiveRoom{
			"room-001": {
				ID:               "room-001",
				Title:            "古风首饰直播竞拍间",
				CoverImage:       "https://images.unsplash.com/photo-1617038220319-276d3cfab638?w=1200&q=80",
				VideoURL:         "https://www.w3schools.com/html/mov_bbb.mp4",
				Status:           "live",
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
				QueueStatus:             "active",
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
				QueueStatus:             "queued",
			},
		},
		sessions: map[string]model.AuctionSession{
			"session-001": {
				ID:                "session-001",
				RoomID:            "room-001",
				ItemID:            "item-001",
				Status:            "bidding",
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
				Status:            "pending",
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
			{ID: "bid-001", SessionID: "session-001", RoomID: "room-001", ItemID: "item-001", UserID: "user-001", BidPrice: 125, RequestID: "req-001", RankAfter: 3, Status: "accepted", CreateTime: now.Add(-90 * time.Second).Format(time.RFC3339)},
			{ID: "bid-002", SessionID: "session-001", RoomID: "room-001", ItemID: "item-001", UserID: "user-002", BidPrice: 130, RequestID: "req-002", RankAfter: 2, Status: "accepted", CreateTime: now.Add(-60 * time.Second).Format(time.RFC3339)},
			{ID: "bid-003", SessionID: "session-001", RoomID: "room-001", ItemID: "item-001", UserID: "user-003", BidPrice: 135, RequestID: "req-003", RankAfter: 1, Status: "accepted", CreateTime: now.Add(-30 * time.Second).Format(time.RFC3339)},
		},
		processedRequests: map[string]model.BidResult{},
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

func nextBidID(count int) string {
	return fmt.Sprintf("bid-%03d", count+1)
}
