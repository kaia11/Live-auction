package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/domain"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/statemachine"
	"auction-live/backend/internal/ws"
)

type OrderHandler struct {
	orderService *service.OrderService
	auditService *service.AuditService
	userService  *service.UserService
	hub          *ws.Hub
}

type updateOrderStatusRequest struct {
	Action string `json:"action"`
}

func NewOrderHandler(orderService *service.OrderService, auditService *service.AuditService, userService *service.UserService, hub *ws.Hub) *OrderHandler {
	return &OrderHandler{orderService: orderService, auditService: auditService, userService: userService, hub: hub}
}

func (h *OrderHandler) ensureAdminAccess(w nethttp.ResponseWriter, r *nethttp.Request) bool {
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

func (h *OrderHandler) ListMyOrders(w nethttp.ResponseWriter, r *nethttp.Request) {
	userID, err := h.userService.GetCurrentUserID(r.Header.Get("Authorization"))
	if err != nil {
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
		return
	}

	api.Success(w, nethttp.StatusOK, h.orderService.ListMyOrders(userID))
}

func (h *OrderHandler) UpdateOrderStatus(w nethttp.ResponseWriter, r *nethttp.Request) {
	if !h.ensureAdminAccess(w, r) {
		return
	}
	operatorID, _ := h.userService.TryGetCurrentUserID(r.Header.Get("Authorization"))

	orderID := r.PathValue("orderId")
	if orderID == "" {
		api.BadRequest(w, "orderId is required")
		return
	}

	var req updateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	event, err := parseOrderEvent(req.Action)
	if err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	order, err := h.orderService.UpdateOrderStatus(orderID, event)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrderNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeNotFound, err.Error())
		case errors.Is(err, service.ErrInvalidOrderState):
			api.Conflict(w, api.CodeInvalidParams, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to update order status")
		}
		return
	}

	h.hub.Publish(order.RoomID, ws.EventAuctionOrderUpdated, order)
	_ = h.auditService.CreateLog("order", "update_status", operatorID, order.RoomID, "order", order.ID, req.Action)
	api.Success(w, nethttp.StatusOK, order)
}

func parseOrderEvent(action string) (statemachine.OrderEvent, error) {
	switch action {
	case string(statemachine.OrderEventMarkPaid):
		return statemachine.OrderEventMarkPaid, nil
	case string(statemachine.OrderEventShip):
		return statemachine.OrderEventShip, nil
	case string(statemachine.OrderEventComplete):
		return statemachine.OrderEventComplete, nil
	case string(statemachine.OrderEventCancel):
		return statemachine.OrderEventCancel, nil
	default:
		return "", service.ErrInvalidOrderState
	}
}
