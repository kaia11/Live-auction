package service

import (
	"errors"
	"testing"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

type stubSessionRepository struct {
	current *model.AuctionSession
	err     error
}

func (r stubSessionRepository) GetCurrentSession(roomID string) (*model.AuctionSession, error) {
	_ = roomID
	return r.current, r.err
}

func (r stubSessionRepository) ListRoomSessions(roomID string) ([]model.AuctionSession, error) {
	_ = roomID
	return nil, nil
}

func (r stubSessionRepository) SaveSession(session model.AuctionSession) error {
	_ = session
	return nil
}

var _ repository.SessionRepository = stubSessionRepository{}

func TestSessionServiceGetCurrentSessionPrefersRepository(t *testing.T) {
	repoSession := &model.AuctionSession{
		ID:            "session-repo",
		RoomID:        "room-001",
		ItemID:        "item-repo",
		Status:        "pending",
		CurrentPrice:  888,
		IncrementStep: 8,
	}

	service := NewSessionService(nil, stubSessionRepository{current: repoSession}, nil, nil)

	got := service.GetCurrentSession("room-001")
	if got.ID != repoSession.ID {
		t.Fatalf("expected repository session %q, got %q", repoSession.ID, got.ID)
	}
	if got.CurrentPrice != repoSession.CurrentPrice {
		t.Fatalf("expected repository price %d, got %d", repoSession.CurrentPrice, got.CurrentPrice)
	}
}

func TestSessionServiceGetCurrentSessionFallsBackToMemory(t *testing.T) {
	store := SharedStore()
	store.mu.RLock()
	expected := store.sessions[store.rooms["room-001"].CurrentSessionID]
	store.mu.RUnlock()

	service := NewSessionService(nil, stubSessionRepository{err: errors.New("mysql offline")}, nil, nil)

	got := service.GetCurrentSession("room-001")
	if got.ID != expected.ID {
		t.Fatalf("expected fallback session %q, got %q", expected.ID, got.ID)
	}
	if got.CurrentPrice != expected.CurrentPrice {
		t.Fatalf("expected fallback price %d, got %d", expected.CurrentPrice, got.CurrentPrice)
	}
}
