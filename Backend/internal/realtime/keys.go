package realtime

import "fmt"

const keyPrefix = "auction"

// RoomCurrentSessionKey stores which session is currently active for a room.
func RoomCurrentSessionKey(roomID string) string {
	return fmt.Sprintf("%s:room:%s:current_session", keyPrefix, roomID)
}

// SessionStateKey stores the hot runtime fields for one auction session.
func SessionStateKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s:state", keyPrefix, sessionID)
}

// SessionRankingKey stores the ranking ZSET for one auction session.
func SessionRankingKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s:ranking", keyPrefix, sessionID)
}

// SessionParticipantsKey stores the participant set for one auction session.
func SessionParticipantsKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s:participants", keyPrefix, sessionID)
}

// RequestDedupKey stores the idempotency record for one bid request.
func RequestDedupKey(requestID string) string {
	return fmt.Sprintf("%s:bid_request:%s", keyPrefix, requestID)
}

// RoomEventVersionKey stores the latest emitted event version for one room.
func RoomEventVersionKey(roomID string) string {
	return fmt.Sprintf("%s:room:%s:event_version", keyPrefix, roomID)
}

// RoomEventStreamKey stores the recent event compensation window for one room.
func RoomEventStreamKey(roomID string) string {
	return fmt.Sprintf("%s:room:%s:event_stream", keyPrefix, roomID)
}

// SessionSettlementLeaseKey stores the short-lived lock for settling one session.
func SessionSettlementLeaseKey(sessionID string) string {
	return fmt.Sprintf("%s:session:%s:settlement_lease", keyPrefix, sessionID)
}
