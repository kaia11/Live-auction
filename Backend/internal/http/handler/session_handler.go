package handler

import (
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type SessionHandler struct {
	sessionService *service.SessionService
	hub            *ws.Hub
}

func NewSessionHandler(sessionService *service.SessionService, hub *ws.Hub) *SessionHandler {
	return &SessionHandler{sessionService: sessionService, hub: hub}
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
