package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"
	"strconv"
	"strings"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type SessionHandler struct {
	sessionService *service.SessionService
	commentService *service.CommentService
	userService    *service.UserService
	hub            *ws.Hub
}

type createRoomCommentRequest struct {
	Content string `json:"content"`
}

func NewSessionHandler(sessionService *service.SessionService, commentService *service.CommentService, userService *service.UserService, hub *ws.Hub) *SessionHandler {
	return &SessionHandler{sessionService: sessionService, commentService: commentService, userService: userService, hub: hub}
}

func (h *SessionHandler) GetCurrentSession(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := r.PathValue("roomId")

	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	api.Success(w, nethttp.StatusOK, h.sessionService.GetCurrentSession(roomID))
}

func (h *SessionHandler) GetRanking(w nethttp.ResponseWriter, r *nethttp.Request) {
	sessionID := r.PathValue("sessionId")

	if sessionID == "" {
		api.BadRequest(w, "sessionId is required")
		return
	}

	api.Success(w, nethttp.StatusOK, h.sessionService.GetRanking(sessionID))
}

func (h *SessionHandler) GetMyStatus(w nethttp.ResponseWriter, r *nethttp.Request) {
	sessionID := r.PathValue("sessionId")
	if sessionID == "" {
		api.BadRequest(w, "sessionId is required")
		return
	}

	userID, err := h.userService.GetCurrentUserID(r.Header.Get("Authorization"))
	if err != nil {
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
		return
	}

	api.Success(w, nethttp.StatusOK, h.sessionService.GetUserStatus(sessionID, userID))
}

func (h *SessionHandler) GetRoomEvents(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := r.PathValue("roomId")
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	sinceVersion, err := parseInt64Query(r, "sinceVersion")
	if err != nil {
		api.BadRequest(w, "sinceVersion must be a valid integer")
		return
	}

	limit, err := parseIntQuery(r, "limit")
	if err != nil {
		api.BadRequest(w, "limit must be a valid integer")
		return
	}

	api.Success(w, nethttp.StatusOK, map[string]any{
		"roomId":        roomID,
		"sinceVersion":  sinceVersion,
		"latestVersion": h.hub.LatestVersion(roomID),
		"events":        h.hub.List(roomID, sinceVersion, limit),
	})
}

func (h *SessionHandler) CreateRoomComment(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := r.PathValue("roomId")
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	var req createRoomCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		api.BadRequest(w, "content is required")
		return
	}

	if len([]rune(content)) > 120 {
		api.BadRequest(w, "content is too long")
		return
	}

	user, err := h.userService.GetCurrentUser(r.Header.Get("Authorization"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUnauthorizedToken):
			api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
		case errors.Is(err, service.ErrUserNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeNotFound, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to resolve current user")
		}
		return
	}

	saved, err := h.commentService.CreateRoomComment(roomID, user.ID, user.Nickname, content)
	if err != nil {
		api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to persist room comment")
		return
	}
	comment := ws.CommentPayload{UserID: saved.UserID, Nickname: saved.Nickname, Content: saved.Content}

	h.hub.Publish(roomID, ws.EventRoomCommentReceived, comment)
	logger.Info("room comment created room_id=%s user_id=%s nickname=%s", roomID, user.ID, user.Nickname)
	api.Success(w, nethttp.StatusOK, comment)
}

func parseInt64Query(r *nethttp.Request, key string) (int64, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}

	return strconv.ParseInt(raw, 10, 64)
}

func parseIntQuery(r *nethttp.Request, key string) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(key))
	if raw == "" {
		return 0, nil
	}

	return strconv.Atoi(raw)
}
