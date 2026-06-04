package repository

import "auction-live/backend/internal/model"

type MemoryStore interface {
	ListRooms() []model.LiveRoom
	ListRoomsByAnchorUserID(anchorUserID string) []model.LiveRoom
	GetRoomDetail(roomID string) model.LiveRoom
	ListRoomItems(roomID string) []model.AuctionItem
	GetItemDetail(roomID string, itemID string) model.AuctionItem
	ListUserOrders(userID string) []model.AuctionOrder
	ListUserBids(userID string) []model.Bid
	GetUserByID(userID string) model.User
	GetUserByUsername(username string) model.User
	SaveUser(user model.User)
}

type MemoryRoomRepository struct {
	store MemoryStore
}

func NewMemoryRoomRepository(store MemoryStore) *MemoryRoomRepository {
	return &MemoryRoomRepository{store: store}
}

func (r *MemoryRoomRepository) ListRooms() ([]model.LiveRoom, error) {
	return r.store.ListRooms(), nil
}

func (r *MemoryRoomRepository) ListRoomsByAnchorUserID(anchorUserID string) ([]model.LiveRoom, error) {
	return r.store.ListRoomsByAnchorUserID(anchorUserID), nil
}

func (r *MemoryRoomRepository) GetRoomDetail(roomID string) (*model.LiveRoom, error) {
	room := r.store.GetRoomDetail(roomID)
	return &room, nil
}

func (r *MemoryRoomRepository) SaveRoom(room model.LiveRoom) error {
	_ = room
	return nil
}

type MemoryItemRepository struct {
	store MemoryStore
}

func NewMemoryItemRepository(store MemoryStore) *MemoryItemRepository {
	return &MemoryItemRepository{store: store}
}

func (r *MemoryItemRepository) ListRoomItems(roomID string) ([]model.AuctionItem, error) {
	return r.store.ListRoomItems(roomID), nil
}

func (r *MemoryItemRepository) GetItemDetail(roomID string, itemID string) (*model.AuctionItem, error) {
	item := r.store.GetItemDetail(roomID, itemID)
	return &item, nil
}

func (r *MemoryItemRepository) SaveItem(item model.AuctionItem) error {
	_ = item
	return nil
}

func (r *MemoryItemRepository) ReplaceRoomQueue(roomID string, itemIDs []string) error {
	_, _ = roomID, itemIDs
	return nil
}

type MemorySessionRepository struct {
	store MemoryStore
}

func NewMemorySessionRepository(store MemoryStore) *MemorySessionRepository {
	return &MemorySessionRepository{store: store}
}

func (r *MemorySessionRepository) GetCurrentSession(roomID string) (*model.AuctionSession, error) {
	room := r.store.GetRoomDetail(roomID)
	if room.CurrentSessionID == "" {
		return &model.AuctionSession{}, nil
	}

	storeWithSessions, ok := r.store.(interface {
		GetCurrentSession(roomID string) model.AuctionSession
	})
	if !ok {
		return &model.AuctionSession{}, nil
	}

	session := storeWithSessions.GetCurrentSession(roomID)
	return &session, nil
}

func (r *MemorySessionRepository) ListRoomSessions(roomID string) ([]model.AuctionSession, error) {
	storeWithSessions, ok := r.store.(interface {
		ListRoomSessions(roomID string) []model.AuctionSession
	})
	if !ok {
		return []model.AuctionSession{}, nil
	}

	return storeWithSessions.ListRoomSessions(roomID), nil
}

func (r *MemorySessionRepository) SaveSession(session model.AuctionSession) error {
	_ = session
	return nil
}

type MemoryOrderRepository struct {
	store MemoryStore
}

func NewMemoryOrderRepository(store MemoryStore) *MemoryOrderRepository {
	return &MemoryOrderRepository{store: store}
}

func (r *MemoryOrderRepository) CreateOrder(order model.AuctionOrder) error {
	_ = order
	return nil
}

func (r *MemoryOrderRepository) ListUserOrders(userID string) ([]model.AuctionOrder, error) {
	return r.store.ListUserOrders(userID), nil
}

func (r *MemoryOrderRepository) UpdateOrder(order model.AuctionOrder) error {
	_ = order
	return nil
}

type MemoryBidRepository struct {
	store MemoryStore
}

func NewMemoryBidRepository(store MemoryStore) *MemoryBidRepository {
	return &MemoryBidRepository{store: store}
}

func (r *MemoryBidRepository) CreateBid(bid model.Bid) error {
	_ = bid
	return nil
}

func (r *MemoryBidRepository) ListUserBids(userID string) ([]model.Bid, error) {
	return r.store.ListUserBids(userID), nil
}

func (r *MemoryBidRepository) ListSessionBids(sessionID string) ([]model.Bid, error) {
	storeWithSessionBids, ok := r.store.(interface {
		ListSessionBids(sessionID string) []model.Bid
	})
	if !ok {
		return []model.Bid{}, nil
	}

	return storeWithSessionBids.ListSessionBids(sessionID), nil
}

type MemoryResultRepository struct{}

func NewMemoryResultRepository() *MemoryResultRepository {
	return &MemoryResultRepository{}
}

func (r *MemoryResultRepository) CreateResult(result model.AuctionResult) error {
	_ = result
	return nil
}

type MemoryCommentRepository struct{}

func NewMemoryCommentRepository() *MemoryCommentRepository {
	return &MemoryCommentRepository{}
}

func (r *MemoryCommentRepository) CreateComment(comment model.RoomComment) error {
	_ = comment
	return nil
}

type MemoryOperationLogRepository struct{}

func NewMemoryOperationLogRepository() *MemoryOperationLogRepository {
	return &MemoryOperationLogRepository{}
}

func (r *MemoryOperationLogRepository) CreateLog(log model.OperationLog) error {
	_ = log
	return nil
}

type MemoryUserRepository struct {
	store MemoryStore
}

func NewMemoryUserRepository(store MemoryStore) *MemoryUserRepository {
	return &MemoryUserRepository{store: store}
}

func (r *MemoryUserRepository) GetByID(userID string) (*model.User, error) {
	user := r.store.GetUserByID(userID)
	return &user, nil
}

func (r *MemoryUserRepository) GetByUsername(username string) (*model.User, error) {
	user := r.store.GetUserByUsername(username)
	return &user, nil
}

func (r *MemoryUserRepository) Create(user model.User) error {
	r.store.SaveUser(user)
	return nil
}

func (r *MemoryUserRepository) UpdatePasswordHash(userID string, passwordHash string) error {
	user := r.store.GetUserByID(userID)
	if user.ID == "" {
		return nil
	}
	user.Password = passwordHash
	r.store.SaveUser(user)
	return nil
}

func (r *MemoryUserRepository) List() ([]model.User, error) {
	storeWithUsers, ok := r.store.(interface {
		ListUsers() []model.User
	})
	if !ok {
		return []model.User{}, nil
	}

	return storeWithUsers.ListUsers(), nil
}
