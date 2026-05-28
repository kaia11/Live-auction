package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type BidHandler struct {
	bidService  *service.BidService
	userService *service.UserService
	hub         *ws.Hub
}

type createBidRequest struct {
	RoomID    string `json:"roomId"`
	SessionID string `json:"sessionId"`
	ItemID    string `json:"itemId"`
	UserID    string `json:"userId"`
	BidPrice  int64  `json:"bidPrice"`
	RequestID string `json:"requestId"`
}

func NewBidHandler(bidService *service.BidService, userService *service.UserService, hub *ws.Hub) *BidHandler {
	return &BidHandler{bidService: bidService, userService: userService, hub: hub}
}

func (h *BidHandler) CreateBid(w nethttp.ResponseWriter, r *nethttp.Request) {
	var req createBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.RoomID == "" || req.SessionID == "" || req.ItemID == "" || req.UserID == "" || req.RequestID == "" {
		api.BadRequest(w, "roomId, sessionId, itemId, userId and requestId are required")
		return
	}

	if req.BidPrice < 0 {
		api.Error(w, nethttp.StatusBadRequest, api.CodeInvalidBidPrice, "bidPrice must not be negative")
		return
	}

	result, settlement, err := h.bidService.CreateBid(service.CreateBidInput{
		RoomID:    req.RoomID,
		SessionID: req.SessionID,
		ItemID:    req.ItemID,
		UserID:    req.UserID,
		BidPrice:  req.BidPrice,
		RequestID: req.RequestID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeRoomNotFound, err.Error())
		case errors.Is(err, service.ErrSessionNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeSessionNotFound, err.Error())
		case errors.Is(err, service.ErrItemNotFound):
			api.Error(w, nethttp.StatusNotFound, api.CodeItemNotFound, err.Error())
		case errors.Is(err, service.ErrSessionNotBidding):
			api.Conflict(w, api.CodeSessionNotBidding, err.Error())
		case errors.Is(err, service.ErrDuplicateBidRequest):
			api.Conflict(w, api.CodeDuplicateBidRequest, err.Error())
		case errors.Is(err, service.ErrInvalidBidPrice), errors.Is(err, service.ErrBidOwnershipMismatch), errors.Is(err, service.ErrUserNotFound):
			api.Error(w, nethttp.StatusBadRequest, api.CodeInvalidBidPrice, err.Error())
		default:
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to create bid")
		}
		return
	}

	h.hub.Publish(req.RoomID, "auction_price_updated", result)
	if settlement != nil {
		h.hub.Publish(req.RoomID, "auction_session_ended", settlement)
		if settlement.Order != nil {
			h.hub.Publish(req.RoomID, "auction_order_created", settlement.Order)
		}
		if settlement.NextSessionID != "" {
			h.hub.Publish(req.RoomID, "auction_session_upcoming", map[string]any{
				"roomId":        settlement.RoomID,
				"nextSessionId": settlement.NextSessionID,
				"nextItemId":    settlement.NextItemID,
			})
		}
	}
	logger.Info(
		"bid accepted room_id=%s session_id=%s item_id=%s user_id=%s price=%d request_id=%s",
		req.RoomID,
		req.SessionID,
		req.ItemID,
		req.UserID,
		result.AcceptedBidPrice,
		req.RequestID,
	)

	api.Success(w, nethttp.StatusOK, result)
}

func (h *BidHandler) ListMyBids(w nethttp.ResponseWriter, r *nethttp.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		resolvedUserID, err := h.userService.GetCurrentUserID(r.Header.Get("Authorization"))
		if err == nil {
			userID = resolvedUserID
		}
	}
	if userID == "" {
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, "missing user identity")
		return
	}

	api.Success(w, nethttp.StatusOK, h.bidService.ListMyBidHistories(userID))
}
