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

	userRepo := repository.NewMemoryUserRepository(store)
	if err := userRepo.Create(model.User{
		ID:       "user-repo-only",
		Username: "repo_only",
		Nickname: "Repo Only",
		Role:     "viewer",
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	delete(store.users, "user-repo-only")

	bidService := NewBidService(nil, repository.NewMemoryBidRepository(store), userRepo, repository.NewMemorySessionRepository(store), nil, nil)
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
