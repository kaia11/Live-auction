package handler

import (
	"errors"
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/service"
)

type RoomHandler struct {
	roomService         *service.RoomService
	liveSnapshotService *service.LiveSnapshotService
	userService         *service.UserService
}

func NewRoomHandler(roomService *service.RoomService, liveSnapshotService *service.LiveSnapshotService, userService *service.UserService) *RoomHandler {
	return &RoomHandler{
		roomService:         roomService,
		liveSnapshotService: liveSnapshotService,
		userService:         userService,
	}
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

func (h *RoomHandler) GetLiveSnapshot(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := r.PathValue("roomId")
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	userID, _ := h.userService.TryGetCurrentUserID(r.Header.Get("Authorization"))

	snapshot, err := h.liveSnapshotService.GetRoomSnapshot(roomID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeRoomNotFound, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to get live snapshot")
		}
		return
	}

	api.Success(w, nethttp.StatusOK, snapshot)
}
