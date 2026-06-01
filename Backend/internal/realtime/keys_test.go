package realtime

import "testing"

func TestRedisKeyLayout(t *testing.T) {
	if got := RoomCurrentSessionKey("room-001"); got != "auction:room:room-001:current_session" {
		t.Fatalf("unexpected room current session key: %s", got)
	}

	if got := SessionStateKey("session-001"); got != "auction:session:session-001:state" {
		t.Fatalf("unexpected session state key: %s", got)
	}

	if got := SessionRankingKey("session-001"); got != "auction:session:session-001:ranking" {
		t.Fatalf("unexpected ranking key: %s", got)
	}

	if got := SessionParticipantsKey("session-001"); got != "auction:session:session-001:participants" {
		t.Fatalf("unexpected participants key: %s", got)
	}

	if got := RequestDedupKey("req-001"); got != "auction:bid_request:req-001" {
		t.Fatalf("unexpected dedup key: %s", got)
	}

	if got := RoomEventVersionKey("room-001"); got != "auction:room:room-001:event_version" {
		t.Fatalf("unexpected event version key: %s", got)
	}

	if got := RoomEventStreamKey("room-001"); got != "auction:room:room-001:event_stream" {
		t.Fatalf("unexpected event stream key: %s", got)
	}
}
