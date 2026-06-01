package realtime

import (
	"encoding/json"
	"strconv"
	"time"

	"auction-live/backend/internal/model"
	"auction-live/backend/internal/ws"
)

type Runtime struct {
	client            *Client
	eventWindowSize   int
	dedupTTLSeconds   int
}

func NewRuntime(client *Client) *Runtime {
	return &Runtime{
		client:          client,
		eventWindowSize: 200,
		dedupTTLSeconds: 3600,
	}
}

func (r *Runtime) Ping() error {
	return r.client.Ping()
}

func (r *Runtime) PublishMessage(roomID string, event string, payload any) (ws.Message, error) {
	version, err := r.client.Incr(RoomEventVersionKey(roomID))
	if err != nil {
		return ws.Message{}, err
	}

	message := ws.Message{
		RoomID:     roomID,
		Event:      event,
		Payload:    payload,
		Version:    version,
		ServerTime: time.Now().Format(time.RFC3339),
	}

	raw, err := json.Marshal(message)
	if err != nil {
		return ws.Message{}, err
	}

	if err := r.client.RPush(RoomEventStreamKey(roomID), string(raw)); err != nil {
		return ws.Message{}, err
	}
	if err := r.client.LTrim(RoomEventStreamKey(roomID), -r.eventWindowSize, -1); err != nil {
		return ws.Message{}, err
	}

	return message, nil
}

func (r *Runtime) List(roomID string, sinceVersion int64, limit int) []ws.Message {
	rawItems, err := r.client.LRange(RoomEventStreamKey(roomID), 0, -1)
	if err != nil {
		return []ws.Message{}
	}

	result := make([]ws.Message, 0)
	for _, raw := range rawItems {
		var message ws.Message
		if json.Unmarshal([]byte(raw), &message) != nil {
			continue
		}
		if message.Version <= sinceVersion {
			continue
		}

		result = append(result, message)
		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result
}

func (r *Runtime) LatestVersion(roomID string) int64 {
	value, ok, err := r.client.Get(RoomEventVersionKey(roomID))
	if err != nil || !ok || value == "" {
		return 0
	}
	version, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return version
}

type SessionState struct {
	SessionID              string
	RoomID                 string
	ItemID                 string
	Status                 string
	CurrentPrice           int64
	LeaderUserID           string
	ParticipantCount       int
	StartPrice             int64
	IncrementStep          int64
	ExtensionSeconds       int
	ExtensionTriggerSeconds int
	CeilingPrice           *int64
	EndTimeUnix            int64
}

func (r *Runtime) SaveSessionState(state SessionState) error {
	ceiling := ""
	if state.CeilingPrice != nil {
		ceiling = strconv.FormatInt(*state.CeilingPrice, 10)
	}

	fields := map[string]string{
		"session_id":                state.SessionID,
		"room_id":                   state.RoomID,
		"item_id":                   state.ItemID,
		"status":                    state.Status,
		"current_price":             strconv.FormatInt(state.CurrentPrice, 10),
		"leader_user_id":            state.LeaderUserID,
		"participant_count":         strconv.Itoa(state.ParticipantCount),
		"start_price":               strconv.FormatInt(state.StartPrice, 10),
		"increment_step":            strconv.FormatInt(state.IncrementStep, 10),
		"extension_seconds":         strconv.Itoa(state.ExtensionSeconds),
		"extension_trigger_seconds": strconv.Itoa(state.ExtensionTriggerSeconds),
		"ceiling_price":             ceiling,
		"end_time_unix":             strconv.FormatInt(state.EndTimeUnix, 10),
	}

	return r.client.HSet(SessionStateKey(state.SessionID), fields)
}

func (r *Runtime) LoadSessionState(sessionID string) (SessionState, bool, error) {
	values, err := r.client.HGetAll(SessionStateKey(sessionID))
	if err != nil {
		return SessionState{}, false, err
	}
	if len(values) == 0 {
		return SessionState{}, false, nil
	}

	state := SessionState{
		SessionID: sessionID,
		RoomID:    values["room_id"],
		ItemID:    values["item_id"],
		Status:    values["status"],
		LeaderUserID: values["leader_user_id"],
	}
	state.CurrentPrice, _ = strconv.ParseInt(values["current_price"], 10, 64)
	participants, _ := strconv.Atoi(values["participant_count"])
	state.ParticipantCount = participants
	state.StartPrice, _ = strconv.ParseInt(values["start_price"], 10, 64)
	state.IncrementStep, _ = strconv.ParseInt(values["increment_step"], 10, 64)
	extSecs, _ := strconv.Atoi(values["extension_seconds"])
	state.ExtensionSeconds = extSecs
	extTrigger, _ := strconv.Atoi(values["extension_trigger_seconds"])
	state.ExtensionTriggerSeconds = extTrigger
	if ceilingText := values["ceiling_price"]; ceilingText != "" {
		parsed, _ := strconv.ParseInt(ceilingText, 10, 64)
		state.CeilingPrice = &parsed
	}
	state.EndTimeUnix, _ = strconv.ParseInt(values["end_time_unix"], 10, 64)

	return state, true, nil
}

func (r *Runtime) SetRoomCurrentSession(roomID, sessionID string) error {
	return r.client.SetEX(RoomCurrentSessionKey(roomID), 24*3600, sessionID)
}

func (r *Runtime) GetRoomCurrentSession(roomID string) (string, bool, error) {
	return r.client.Get(RoomCurrentSessionKey(roomID))
}

func (r *Runtime) ReplaceRanking(sessionID string, entries []model.RankingEntry) error {
	rankingKey := SessionRankingKey(sessionID)
	participantKey := SessionParticipantsKey(sessionID)

	if err := r.client.Del(rankingKey, participantKey); err != nil {
		return err
	}
	for _, entry := range entries {
		if err := r.client.ZAdd(rankingKey, entry.HighestBid, entry.UserID); err != nil {
			return err
		}
		if err := r.client.SAdd(participantKey, entry.UserID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runtime) GetTopRanking(sessionID string, limit int, users map[string]model.User) ([]model.RankingEntry, error) {
	if limit <= 0 {
		limit = 3
	}
	rows, err := r.client.ZRevRangeWithScores(SessionRankingKey(sessionID), 0, limit-1)
	if err != nil {
		return nil, err
	}

	result := make([]model.RankingEntry, 0, len(rows))
	for index, row := range rows {
		user := users[row.Member]
		result = append(result, model.RankingEntry{
			UserID:     row.Member,
			Nickname:   user.Nickname,
			Avatar:     user.Avatar,
			Rank:       index + 1,
			HighestBid: row.Score,
		})
	}
	return result, nil
}

func (r *Runtime) GetUserRankingEntry(sessionID, userID string, users map[string]model.User) (model.RankingEntry, bool, error) {
	rank, ok, err := r.client.ZRevRank(SessionRankingKey(sessionID), userID)
	if err != nil || !ok {
		return model.RankingEntry{}, false, err
	}
	score, ok, err := r.client.ZScore(SessionRankingKey(sessionID), userID)
	if err != nil || !ok {
		return model.RankingEntry{}, false, err
	}
	user := users[userID]
	return model.RankingEntry{
		UserID:     userID,
		Nickname:   user.Nickname,
		Avatar:     user.Avatar,
		Rank:       int(rank) + 1,
		HighestBid: score,
	}, true, nil
}

type AtomicBidInput struct {
	SessionID string
	UserID    string
	BidPrice  int64
	RequestID string
	NowUnix   int64
}

type AtomicBidResult struct {
	OK               bool   `json:"ok"`
	Code             string `json:"code,omitempty"`
	AcceptedBidPrice int64  `json:"accepted_bid_price"`
	CurrentPrice     int64  `json:"current_price"`
	ParticipantCount int    `json:"participant_count"`
	Rank             int    `json:"rank"`
	CeilingReached   bool   `json:"ceiling_reached"`
	ExtensionApplied bool   `json:"extension_applied"`
	EndTimeUnix      int64  `json:"end_time_unix"`
	NextMinimumBid   int64  `json:"next_minimum_bid"`
}

func (r *Runtime) RunAtomicBid(input AtomicBidInput) (AtomicBidResult, error) {
	response, err := r.client.Eval(
		AtomicBidScript,
		[]string{
			SessionStateKey(input.SessionID),
			SessionRankingKey(input.SessionID),
			SessionParticipantsKey(input.SessionID),
			RequestDedupKey(input.RequestID),
		},
		[]string{
			input.RequestID,
			input.UserID,
			strconv.FormatInt(input.BidPrice, 10),
			strconv.FormatInt(input.NowUnix, 10),
			strconv.Itoa(r.dedupTTLSeconds),
		},
	)
	if err != nil {
		return AtomicBidResult{}, err
	}

	var result AtomicBidResult
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		return AtomicBidResult{}, err
	}

	return result, nil
}
