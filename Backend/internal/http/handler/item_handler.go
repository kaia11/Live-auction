package handler

import (
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/service"
)

type ItemHandler struct {
	itemService *service.ItemService
}

func NewItemHandler(itemService *service.ItemService) *ItemHandler {
	return &ItemHandler{itemService: itemService}
}

func (h *ItemHandler) ListRoomItems(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := r.PathValue("roomId")

	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	api.Success(w, nethttp.StatusOK, h.itemService.ListRoomItems(roomID))
}

func (h *ItemHandler) GetItemDetail(w nethttp.ResponseWriter, r *nethttp.Request) {
	roomID := r.PathValue("roomId")
	itemID := r.PathValue("itemId")

	if roomID == "" || itemID == "" {
		api.BadRequest(w, "roomId and itemId are required")
		return
	}

	api.Success(w, nethttp.StatusOK, h.itemService.GetItemDetail(roomID, itemID))
}
