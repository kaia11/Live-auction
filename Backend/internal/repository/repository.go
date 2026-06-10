package repository

import "auction-live/backend/internal/model"

type RoomRepository interface {
	ListRooms() ([]model.LiveRoom, error)
	ListRoomsByAnchorUserID(anchorUserID string) ([]model.LiveRoom, error)
	GetRoomDetail(roomID string) (*model.LiveRoom, error)
	SaveRoom(room model.LiveRoom) error
}

type ItemRepository interface {
	ListRoomItems(roomID string) ([]model.AuctionItem, error)
	GetItemDetail(roomID string, itemID string) (*model.AuctionItem, error)
	SaveItem(item model.AuctionItem) error
	ReplaceRoomQueue(roomID string, itemIDs []string) error
}

type SessionRepository interface {
	GetCurrentSession(roomID string) (*model.AuctionSession, error)
	ListRoomSessions(roomID string) ([]model.AuctionSession, error)
	SaveSession(session model.AuctionSession) error
}

type BidRepository interface {
	CreateBid(bid model.Bid) error
	ListUserBids(userID string) ([]model.Bid, error)
	ListSessionBids(sessionID string) ([]model.Bid, error)
}

type OrderRepository interface {
	CreateOrder(order model.AuctionOrder) error
	GetOrderByID(orderID string) (*model.AuctionOrder, error)
	ListAllOrders() ([]model.AuctionOrder, error)
	ListUserOrders(userID string) ([]model.AuctionOrder, error)
	UpdateOrder(order model.AuctionOrder) error
}

type RankingRepository interface {
	GetTopRanks(sessionID string, limit int) ([]model.RankingEntry, error)
	GetUserRank(sessionID string, userID string) (*model.RankingEntry, error)
}

type ResultRepository interface {
	CreateResult(result model.AuctionResult) error
}

type UserRepository interface {
	GetByID(userID string) (*model.User, error)
	GetByUsername(username string) (*model.User, error)
	Create(user model.User) error
	UpdatePasswordHash(userID string, passwordHash string) error
	List() ([]model.User, error)
}

type CommentRepository interface {
	CreateComment(comment model.RoomComment) error
}

type OperationLogRepository interface {
	CreateLog(log model.OperationLog) error
}
