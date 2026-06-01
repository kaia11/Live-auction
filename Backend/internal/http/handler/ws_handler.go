package handler

import (
	"errors"
	"io"
	nethttp "net/http"
	"strings"
	"time"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/monitoring"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type WebSocketHandler struct {
	userService *service.UserService
	roomService *service.RoomService
	hub         *ws.Hub
	metrics     *monitoring.Metrics
}

func NewWebSocketHandler(userService *service.UserService, roomService *service.RoomService, hub *ws.Hub, metrics *monitoring.Metrics) *WebSocketHandler {
	return &WebSocketHandler{
		userService: userService,
		roomService: roomService,
		hub:         hub,
		metrics:     metrics,
	}
}

func (h *WebSocketHandler) ServeRoomStream(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := strings.TrimSpace(r.URL.Query().Get("roomId"))
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	if room := h.roomService.GetRoomDetail(roomID); room.ID == "" {
		api.Error(w, nethttp.StatusNotFound, api.CodeRoomNotFound, "room not found")
		return
	}

	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = r.Header.Get("Authorization")
	}
	if token != "" {
		if _, err := h.userService.GetCurrentUserID(token); err != nil {
			api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
			return
		}
	}

	conn, err := ws.Upgrade(w, r)
	if err != nil {
		logger.Error("websocket upgrade failed room_id=%s error=%v", roomID, err)
		return
	}
	defer conn.Close()

	client := ws.NewClient(roomID, 64)
	unregister := h.hub.Register(roomID, client)
	if h.metrics != nil {
		h.metrics.RecordWSConnect()
	}
	defer unregister()
	defer func() {
		if h.metrics != nil {
			h.metrics.RecordWSDisconnect()
		}
	}()
	defer client.Close()

	logger.Info("websocket connected room_id=%s", roomID)

	if err := conn.SetReadDeadline(time.Now().Add(70 * time.Second)); err != nil {
		logger.Error("websocket set read deadline failed room_id=%s error=%v", roomID, err)
		return
	}

	readErrCh := make(chan error, 1)
	go func() {
		readErrCh <- conn.ReadLoop(
			func() {
				_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
			},
			func(payload []byte) error {
				_ = conn.SetReadDeadline(time.Now().Add(70 * time.Second))
				return conn.WritePong(payload)
			},
		)
	}()

	pingTicker := time.NewTicker(25 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case message, ok := <-client.Messages():
			if !ok {
				return
			}
			if err := conn.WriteJSON(message); err != nil {
				logger.Error("websocket write failed room_id=%s error=%v", roomID, err)
				return
			}
		case <-pingTicker.C:
			if err := conn.WritePing(); err != nil {
				logger.Error("websocket ping failed room_id=%s error=%v", roomID, err)
				return
			}
		case err := <-readErrCh:
			if err != nil && !errors.Is(err, io.EOF) {
				logger.Error("websocket read loop stopped room_id=%s error=%v", roomID, err)
			}
			logger.Info("websocket disconnected room_id=%s", roomID)
			return
		}
	}
}
