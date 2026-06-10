package service

import (
	"errors"
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

	bidService := NewBidService(nil, false, repository.NewMemoryBidRepository(store), repository.NewMemoryRoomRepository(store), repository.NewMemoryItemRepository(store), userRepo, repository.NewMemorySessionRepository(store), nil, nil)
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

	bidService := NewBidService(nil, false, repository.NewMemoryBidRepository(store), roomRepo, itemRepo, userRepo, sessionRepo, nil, nil)
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

func TestBidServiceCreateBidRejectsCurrentLeader(t *testing.T) {
	store := newMemoryStore()
	bidService := NewBidService(nil, false, repository.NewMemoryBidRepository(store), repository.NewMemoryRoomRepository(store), repository.NewMemoryItemRepository(store), repository.NewMemoryUserRepository(store), repository.NewMemorySessionRepository(store), nil, nil)
	bidService.store = store
	bidService.engine = NewAuctionEngine(store, nil, nil, nil, nil, nil, nil)

	_, _, err := bidService.CreateBid(CreateBidInput{
		RoomID:    "room-001",
		SessionID: "session-001",
		ItemID:    "item-001",
		UserID:    "user-003",
		BidPrice:  140,
		RequestID: "req-leading-user",
	})
	if !errors.Is(err, ErrAlreadyLeadingBid) {
		t.Fatalf("expected ErrAlreadyLeadingBid, got %v", err)
	}
}

func TestBidServiceCreateBidAllowsCurrentLeaderWhenRepeatBidEnabled(t *testing.T) {
	store := newMemoryStore()
	bidService := NewBidService(nil, true, repository.NewMemoryBidRepository(store), repository.NewMemoryRoomRepository(store), repository.NewMemoryItemRepository(store), repository.NewMemoryUserRepository(store), repository.NewMemorySessionRepository(store), nil, nil)
	bidService.store = store
	bidService.engine = NewAuctionEngine(store, nil, nil, nil, nil, nil, nil)

	result, _, err := bidService.CreateBid(CreateBidInput{
		RoomID:    "room-001",
		SessionID: "session-001",
		ItemID:    "item-001",
		UserID:    "user-003",
		BidPrice:  140,
		RequestID: "req-leading-user-repeat",
	})
	if err != nil {
		t.Fatalf("CreateBid() error = %v", err)
	}
	if result.UserID != "user-003" {
		t.Fatalf("expected user-003, got %s", result.UserID)
	}
}
