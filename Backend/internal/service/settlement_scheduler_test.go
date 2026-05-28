package service

import (
	"testing"
	"time"

	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/model"
	"auction-live/backend/internal/ws"
)

func TestSettlementSchedulerScanOnceSettlesExpiredSession(t *testing.T) {
	store := newMemoryStore()
	store.sessions["session-001"] = forceSessionEndTime(store.sessions["session-001"], time.Now().Add(-time.Minute).Format(time.RFC3339))
	scheduler := NewSettlementScheduler(store, ws.NewHub(), time.Second)

	outcomes := scheduler.ScanOnce()
	if len(outcomes) != 1 {
		t.Fatalf("expected 1 outcome, got %d", len(outcomes))
	}
	if outcomes[0].Status != domain.SessionStateEndedSold {
		t.Fatalf("unexpected outcome status %q", outcomes[0].Status)
	}
}

func forceSessionEndTime(session model.AuctionSession, endTime string) model.AuctionSession {
	session.EndTime = endTime
	return session
}
