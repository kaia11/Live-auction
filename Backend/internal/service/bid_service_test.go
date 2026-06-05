package service

import (
	"testing"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

func TestBidServiceCreateBidAcceptsRepoBackedUserOutsideSharedStore(t *testing.T) {
	store := newMemoryStore()
	store.users = map[string]model.User{}
	store.userIDsByUsername = map[string]string{}

	repoStore := newMemoryStore()
	userRepo := repository.NewMemoryUserRepository(repoStore)
	if err := userRepo.Create(model.User{
		ID:       "user-repo-only",
		Username: "repo_only",
		Nickname: "Repo Only",
		Role:     "viewer",
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	bidService := NewBidService(nil, repository.NewMemoryBidRepository(store), repository.NewMemoryRoomRepository(store), repository.NewMemoryItemRepository(store), userRepo, repository.NewMemorySessionRepository(store), nil, nil)
	bidService.store = store
	bidService.engine = NewAuctionEngine(store, nil, nil, nil, nil, nil, nil)

	result, _, err := bidService.CreateBid(CreateBidInput{
		RoomID:    "room-001",
		SessionID: "session-001",
		ItemID:    "item-001",
		UserID:    "user-repo-only",
		BidPrice:  140,
		RequestID: "req-repo-user",
	})
	if err != nil {
		t.Fatalf("CreateBid() error = %v", err)
	}
	if result.UserID != "user-repo-only" {
		t.Fatalf("expected user-repo-only, got %s", result.UserID)
	}
}

func TestBidServiceCreateBidUsesRepositoryContextForRoomSessionItem(t *testing.T) {
	store := newMemoryStore()
	store.rooms["room-001"] = model.LiveRoom{}
	store.sessions["session-001"] = model.AuctionSession{}
	store.items["item-001"] = model.AuctionItem{}

	roomRepo := repository.NewMemoryRoomRepository(newMemoryStore())
	itemRepo := repository.NewMemoryItemRepository(newMemoryStore())
	sessionRepo := repository.NewMemorySessionRepository(newMemoryStore())
	userRepo := repository.NewMemoryUserRepository(newMemoryStore())

	bidService := NewBidService(nil, repository.NewMemoryBidRepository(store), roomRepo, itemRepo, userRepo, sessionRepo, nil, nil)
	bidService.store = store
	bidService.engine = NewAuctionEngine(store, nil, nil, nil, nil, nil, nil)

	result, _, err := bidService.CreateBid(CreateBidInput{
		RoomID:    "room-001",
		SessionID: "session-001",
		ItemID:    "item-001",
		UserID:    "user-001",
		BidPrice:  140,
		RequestID: "req-repo-context",
	})
	if err != nil {
		t.Fatalf("CreateBid() error = %v", err)
	}
	if result.ItemID != "item-001" {
		t.Fatalf("expected item-001, got %s", result.ItemID)
	}
}
