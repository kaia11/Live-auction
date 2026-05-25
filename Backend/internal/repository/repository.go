package repository

import "auction-live/backend/internal/model"

type RoomRepository interface {
	ListRooms() ([]model.LiveRoom, error)
	GetRoomDetail(roomID string) (*model.LiveRoom, error)
}

type ItemRepository interface {
	ListRoomItems(roomID string) ([]model.AuctionItem, error)
	GetItemDetail(roomID string, itemID string) (*model.AuctionItem, error)
}

type SessionRepository interface {
	GetCurrentSession(roomID string) (*model.AuctionSession, error)
	ListRoomSessions(roomID string) ([]model.AuctionSession, error)
}

type BidRepository interface {
	CreateBid(bid model.Bid) error
	ListUserBids(userID string) ([]model.Bid, error)
}

type RankingRepository interface {
	GetTopRanks(sessionID string, limit int) ([]model.RankingEntry, error)
	GetUserRank(sessionID string, userID string) (*model.RankingEntry, error)
}

type ResultRepository interface {
	CreateResult(result model.AuctionResult) error
}
