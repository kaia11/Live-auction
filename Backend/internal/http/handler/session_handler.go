package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"
	"strings"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type SessionHandler struct {
	sessionService *service.SessionService
	userService    *service.UserService
	hub            *ws.Hub
}

type createRoomCommentRequest struct {
	Content string `json:"content"`
}

func NewSessionHandler(sessionService *service.SessionService, userService *service.UserService, hub *ws.Hub) *SessionHandler {
	return &SessionHandler{sessionService: sessionService, userService: userService, hub: hub}
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
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "user-001"
	}

	if sessionID == "" {
		api.BadRequest(w, "sessionId is required")
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

	api.Success(w, nethttp.StatusOK, h.hub.List(roomID))
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

	comment := ws.CommentPayload{
		UserID:   user.ID,
		Nickname: user.Nickname,
		Content:  content,
	}

	h.hub.Publish(roomID, "room_comment_received", comment)
	api.Success(w, nethttp.StatusOK, comment)
}
