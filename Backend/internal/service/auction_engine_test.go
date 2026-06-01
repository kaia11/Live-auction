package service

import (
	"testing"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
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

func withLeader(session model.AuctionSession, leader string) model.AuctionSession {
	session.LeaderUserID = leader
	return session
}
