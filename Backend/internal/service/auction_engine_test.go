package service

import (
	"errors"
	"testing"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/repository"
)

func TestAuctionEngineSettleSessionCreatesOrderAndPreparesNext(t *testing.T) {
	store := newMemoryStore()
	engine := NewAuctionEngine(store, nil, nil, nil, nil, nil, nil)

	outcome, err := engine.SettleSessionLocked("session-001")
	if err != nil {
		t.Fatalf("SettleSessionLocked() error = %v", err)
	}

	if outcome.Status != domain.SessionStateEndedSold {
		t.Fatalf("unexpected session status %q", outcome.Status)
	}
	if outcome.Order == nil {
		t.Fatal("expected order to be created")
	}
	if outcome.NextSessionID != "session-002" {
		t.Fatalf("unexpected next session %q", outcome.NextSessionID)
	}
	if store.items["item-001"].QueueStatus != domain.QueueStateFinished {
		t.Fatalf("unexpected current item queue status %q", store.items["item-001"].QueueStatus)
	}
	if store.items["item-002"].QueueStatus != domain.QueueStateUpcoming {
		t.Fatalf("unexpected next item queue status %q", store.items["item-002"].QueueStatus)
	}
}

func TestAuctionEngineSettleSessionNoBidEndsPassed(t *testing.T) {
	store := newMemoryStore()
	store.sessions["session-001"] = withLeader(store.sessions["session-001"], "")
	engine := NewAuctionEngine(store, nil, nil, nil, nil, nil, nil)

	outcome, err := engine.SettleSessionLocked("session-001")
	if err != nil {
		t.Fatalf("SettleSessionLocked() error = %v", err)
	}

	if outcome.Status != domain.SessionStateEndedPassed {
		t.Fatalf("unexpected session status %q", outcome.Status)
	}
	if outcome.Order != nil {
		t.Fatal("did not expect order for passed session")
	}
}

func TestAuctionEngineStartSessionRequiresRoomLive(t *testing.T) {
	store := newMemoryStore()
	store.rooms["room-001"] = withRoomStatus(store.rooms["room-001"], domain.RoomStatusOffline)
	store.items["item-001"] = withQueueStatus(store.items["item-001"], domain.QueueStateUpcoming)
	store.sessions["session-001"] = withSessionStatus(store.sessions["session-001"], domain.SessionStatePending)
	engine := NewAuctionEngine(store, nil, nil, nil, nil, nil, nil)

	_, err := engine.StartSessionLocked("session-001")
	if !errors.Is(err, ErrRoomNotLive) {
		t.Fatalf("expected ErrRoomNotLive, got %v", err)
	}
}

func TestAdminServiceStopRoomLiveRejectsActiveSession(t *testing.T) {
	store := newMemoryStore()
	adminService := NewAdminService(nil, repository.NewMemoryRoomRepository(store), repository.NewMemoryItemRepository(store), repository.NewMemorySessionRepository(store), repository.NewMemoryBidRepository(store), nil, nil)
	adminService.store = store

	_, err := adminService.StopRoomLive("room-001")
	if !errors.Is(err, ErrRoomHasActiveSession) {
		t.Fatalf("expected ErrRoomHasActiveSession, got %v", err)
	}
}

func TestAdminServiceStartRoomLiveMarksRoomLive(t *testing.T) {
	store := newMemoryStore()
	store.rooms["room-001"] = withRoomStatus(store.rooms["room-001"], domain.RoomStatusOffline)
	adminService := NewAdminService(nil, repository.NewMemoryRoomRepository(store), repository.NewMemoryItemRepository(store), repository.NewMemorySessionRepository(store), repository.NewMemoryBidRepository(store), nil, nil)
	adminService.store = store

	result, err := adminService.StartRoomLive("room-001")
	if err != nil {
		t.Fatalf("StartRoomLive() error = %v", err)
	}

	if result["status"] != domain.RoomStatusLive {
		t.Fatalf("expected status live, got %v", result["status"])
	}
	if store.rooms["room-001"].Status != domain.RoomStatusLive {
		t.Fatalf("expected room status live, got %q", store.rooms["room-001"].Status)
	}
}

func withLeader(session model.AuctionSession, leader string) model.AuctionSession {
	session.LeaderUserID = leader
	return session
}

func withSessionStatus(session model.AuctionSession, status string) model.AuctionSession {
	session.Status = status
	return session
}

func withQueueStatus(item model.AuctionItem, status string) model.AuctionItem {
	item.QueueStatus = status
	return item
}

func withRoomStatus(room model.LiveRoom, status string) model.LiveRoom {
	room.Status = status
	return room
}
