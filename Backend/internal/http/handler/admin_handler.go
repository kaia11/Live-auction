package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type AdminHandler struct {
	adminService *service.AdminService
	userService  *service.UserService
	hub          *ws.Hub
}

type reorderQueueRequest struct {
	ItemIDs []string `json:"itemIds"`
}

type createItemRequest struct {
	Title                   string `json:"title"`
	CoverImage              string `json:"coverImage"`
	Description             string `json:"description"`
	StartPrice              int64  `json:"startPrice"`
	IncrementStep           int64  `json:"incrementStep"`
	CeilingPrice            *int64 `json:"ceilingPrice"`
	DurationSeconds         int    `json:"durationSeconds"`
	ExtensionSeconds        int    `json:"extensionSeconds"`
	ExtensionTriggerSeconds int    `json:"extensionTriggerSeconds"`
}

type updateItemRequest struct {
	Title                   *string `json:"title"`
	CoverImage              *string `json:"coverImage"`
	Description             *string `json:"description"`
	StartPrice              *int64  `json:"startPrice"`
	IncrementStep           *int64  `json:"incrementStep"`
	CeilingPrice            *int64  `json:"ceilingPrice"`
	DurationSeconds         *int    `json:"durationSeconds"`
	ExtensionSeconds        *int    `json:"extensionSeconds"`
	ExtensionTriggerSeconds *int    `json:"extensionTriggerSeconds"`
}

func NewAdminHandler(adminService *service.AdminService, userService *service.UserService, hub *ws.Hub) *AdminHandler {
	return &AdminHandler{adminService: adminService, userService: userService, hub: hub}
}

func (h *AdminHandler) ensureAdminAccess(w nethttp.ResponseWriter, r *nethttp.Request) bool {
	_, err := h.userService.RequireAnyRole(r.Header.Get("Authorization"), domain.UserRoleAnchor, domain.UserRoleAdmin)
	if err == nil {
		return true
	}

	switch {
	case errors.Is(err, service.ErrUnauthorizedToken):
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
	case errors.Is(err, service.ErrForbiddenRole):
		api.Error(w, nethttp.StatusForbidden, api.CodeForbidden, err.Error())
	default:
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
	}

	return false
}

func (h *AdminHandler) CreateItem(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	roomID := r.PathValue("roomId")
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	var req createItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	item, meta, err := h.adminService.CreateItem(roomID, service.CreateItemInput{
		Title:                   req.Title,
		CoverImage:              req.CoverImage,
		Description:             req.Description,
		StartPrice:              req.StartPrice,
		IncrementStep:           req.IncrementStep,
		CeilingPrice:            req.CeilingPrice,
		DurationSeconds:         req.DurationSeconds,
		ExtensionSeconds:        req.ExtensionSeconds,
		ExtensionTriggerSeconds: req.ExtensionTriggerSeconds,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeRoomNotFound, err.Error())
		default:
			api.BadRequest(w, err.Error())
		}
		return
	}

	h.hub.Publish(roomID, ws.EventRoomItemQueueUpdated, meta)
	api.Created(w, item)
}

func (h *AdminHandler) UpdateItem(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	itemID := r.PathValue("itemId")
	if itemID == "" {
		api.BadRequest(w, "itemId is required")
		return
	}

	var req updateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	item, err := h.adminService.UpdateItem(itemID, service.UpdateItemInput{
		Title:                   req.Title,
		CoverImage:              req.CoverImage,
		Description:             req.Description,
		StartPrice:              req.StartPrice,
		IncrementStep:           req.IncrementStep,
		CeilingPrice:            req.CeilingPrice,
		DurationSeconds:         req.DurationSeconds,
		ExtensionSeconds:        req.ExtensionSeconds,
		ExtensionTriggerSeconds: req.ExtensionTriggerSeconds,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrItemNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeItemNotFound, err.Error())
		default:
			api.Conflict(w, api.CodeInvalidParams, err.Error())
		}
		return
	}

	h.hub.Publish(item.RoomID, ws.EventRoomItemQueueUpdated, map[string]any{
		"itemId":      item.ID,
		"queueStatus": item.QueueStatus,
		"title":       item.Title,
	})
	api.Success(w, nethttp.StatusOK, item)
}

func (h *AdminHandler) ReorderQueue(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	roomID := r.PathValue("roomId")
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	var req reorderQueueRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	result, err := h.adminService.ReorderQueue(roomID, req.ItemIDs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeRoomNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidQueueOrder):
			api.Conflict(w, api.CodeInvalidParams, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to reorder queue")
		}
		return
	}

	api.Success(w, nethttp.StatusOK, result)
	h.hub.Publish(roomID, ws.EventRoomItemQueueUpdated, result)
}

func (h *AdminHandler) ActivateNextItem(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	roomID := r.PathValue("roomId")
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	result, err := h.adminService.ActivateNextItem(roomID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeRoomNotFound, err.Error())
		case errors.Is(err, service.ErrQueueExhausted):
			api.Conflict(w, api.CodeInvalidParams, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to activate next item")
		}
		return
	}

	api.Success(w, nethttp.StatusOK, result)
	h.hub.Publish(roomID, ws.EventAuctionSessionUpcoming, result)
}

func (h *AdminHandler) StartSession(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	sessionID := r.PathValue("sessionId")
	if sessionID == "" {
		api.BadRequest(w, "sessionId is required")
		return
	}

	result, err := h.adminService.StartSession(sessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeSessionNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidSessionState):
			api.Conflict(w, api.CodeSessionNotBidding, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to start session")
		}
		return
	}

	roomID, _ := result["roomId"].(string)
	if roomID != "" {
		h.hub.Publish(roomID, ws.EventAuctionSessionActivated, result)
	}
	api.Success(w, nethttp.StatusOK, result)
}

func (h *AdminHandler) CancelSession(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	sessionID := r.PathValue("sessionId")
	if sessionID == "" {
		api.BadRequest(w, "sessionId is required")
		return
	}

	result, err := h.adminService.CancelSession(sessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeSessionNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidSessionState):
			api.Conflict(w, api.CodeSessionNotBidding, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to cancel session")
		}
		return
	}

	roomID, _ := result["roomId"].(string)
	if roomID != "" {
		h.hub.Publish(roomID, ws.EventAuctionSessionEnded, result)
	}
	api.Success(w, nethttp.StatusOK, result)
}

func (h *AdminHandler) SettleSession(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	sessionID := r.PathValue("sessionId")
	if sessionID == "" {
		api.BadRequest(w, "sessionId is required")
		return
	}

	result, err := h.adminService.SettleSession(sessionID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSessionNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeSessionNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidSessionState):
			api.Conflict(w, api.CodeSessionNotBidding, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to settle session")
		}
		return
	}

	if result.RoomID != "" {
		h.hub.Publish(result.RoomID, ws.EventAuctionSessionEnded, result)
		if result.Order != nil {
			h.hub.Publish(result.RoomID, ws.EventAuctionOrderCreated, result.Order)
		}
		if result.NextSessionID != "" {
			h.hub.Publish(result.RoomID, ws.EventAuctionSessionUpcoming, map[string]any{
				"roomId":        result.RoomID,
				"nextSessionId": result.NextSessionID,
				"nextItemId":    result.NextItemID,
			})
		}
	}

	api.Success(w, nethttp.StatusOK, result)
}

func (h *AdminHandler) ListRoomSessions(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	roomID := r.PathValue("roomId")
	if roomID == "" {
		api.BadRequest(w, "roomId is required")
		return
	}

	api.Success(w, nethttp.StatusOK, h.adminService.ListRoomSessions(roomID))
}

func (h *AdminHandler) ListOrders(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	api.Success(w, nethttp.StatusOK, h.adminService.ListOrders())
}

func (h *AdminHandler) GetStatsOverview(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	api.Success(w, nethttp.StatusOK, h.adminService.GetStatsOverview())
}

func (h *AdminHandler) GetStatsTimeline(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}

	api.Success(w, nethttp.StatusOK, h.adminService.GetStatsTimeline())
}
