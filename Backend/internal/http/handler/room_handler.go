package handler

import (
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/service"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

func (h *RoomHandler) ListRooms(w nethttp.ResponseWriter, r *nethttp.Request) {
	api.Success(w, nethttp.StatusOK, h.roomService.ListRooms())
}

func (h *RoomHandler) GetRoomDetail(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := r.PathValue("roomId")

	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	api.Success(w, nethttp.StatusOK, h.roomService.GetRoomDetail(roomID))
}
