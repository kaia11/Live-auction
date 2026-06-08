package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/monitoring"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type BidHandler struct {
	bidService  *service.BidService
	userService *service.UserService
	hub         *ws.Hub
	metrics     *monitoring.Metrics
}

type createBidRequest struct {
	RoomID    string `json:"roomId"`
	SessionID string `json:"sessionId"`
	ItemID    string `json:"itemId"`
	UserID    string `json:"userId"`
	BidPrice  int64  `json:"bidPrice"`
	RequestID string `json:"requestId"`
}

func NewBidHandler(bidService *service.BidService, userService *service.UserService, hub *ws.Hub, metrics *monitoring.Metrics) *BidHandler {
	return &BidHandler{bidService: bidService, userService: userService, hub: hub, metrics: metrics}
}

func (h *BidHandler) CreateBid(w nethttp.ResponseWriter, r *nethttp.Request) {
	if h.metrics != nil {
		h.metrics.RecordBidAttempt()
	}
	userID, err := h.userService.GetCurrentUserID(r.Header.Get("Authorization"))
	if err != nil {
		if h.metrics != nil {
			h.metrics.RecordBidFailure("unauthorized")
		}
		api.Error(w, nethttp.StatusUnauthorized, api.CodeUnauthorized, err.Error())
		return
	}

	var req createBidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if h.metrics != nil {
			h.metrics.RecordBidFailure("invalid_body")
		}
		api.BadRequest(w, "invalid request body")
		return
	}

	if req.RoomID == "" || req.SessionID == "" || req.ItemID == "" || req.RequestID == "" {
		if h.metrics != nil {
			h.metrics.RecordBidFailure("missing_params")
		}
		api.BadRequest(w, "roomId, sessionId, itemId and requestId are required")
		return
	}
	if req.UserID != "" && req.UserID != userID {
		if h.metrics != nil {
			h.metrics.RecordBidFailure("user_mismatch")
		}
		api.Error(w, nethttp.StatusForbidden, api.CodeForbidden, "request userId does not match token user")
		return
	}

	if req.BidPrice < 0 {
		if h.metrics != nil {
			h.metrics.RecordBidFailure("negative_price")
		}
		api.Error(w, nethttp.StatusBadRequest, api.CodeInvalidBidPrice, "bidPrice must not be negative")
		return
	}

	result, settlement, err := h.bidService.CreateBid(service.CreateBidInput{
		RoomID:    req.RoomID,
		SessionID: req.SessionID,
		ItemID:    req.ItemID,
		UserID:    userID,
		BidPrice:  req.BidPrice,
		RequestID: req.RequestID,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomNotFound):
			if h.metrics != nil {
				h.metrics.RecordBidFailure("room_not_found")
			}
			api.Error(w, nethttp.StatusNotFound, api.CodeRoomNotFound, err.Error())
		case errors.Is(err, service.ErrSessionNotFound):
			if h.metrics != nil {
				h.metrics.RecordBidFailure("session_not_found")
			}
			api.Error(w, nethttp.StatusNotFound, api.CodeSessionNotFound, err.Error())
		case errors.Is(err, service.ErrItemNotFound):
			if h.metrics != nil {
				h.metrics.RecordBidFailure("item_not_found")
			}
			api.Error(w, nethttp.StatusNotFound, api.CodeItemNotFound, err.Error())
		case errors.Is(err, service.ErrSessionNotBidding):
			if h.metrics != nil {
				h.metrics.RecordBidFailure("session_not_bidding")
			}
			api.Conflict(w, api.CodeSessionNotBidding, err.Error())
		case errors.Is(err, service.ErrDuplicateBidRequest):
			if h.metrics != nil {
				h.metrics.RecordBidFailure("duplicate_request")
			}
			api.Conflict(w, api.CodeDuplicateBidRequest, err.Error())
		case errors.Is(err, service.ErrAlreadyLeadingBid):
			if h.metrics != nil {
				h.metrics.RecordBidFailure("already_leading")
			}
			api.Conflict(w, api.CodeInvalidBidPrice, err.Error())
		case errors.Is(err, service.ErrInvalidBidPrice), errors.Is(err, service.ErrBidOwnershipMismatch), errors.Is(err, service.ErrUserNotFound):
			if h.metrics != nil {
				h.metrics.RecordBidFailure("invalid_bid")
			}
			api.Error(w, nethttp.StatusBadRequest, api.CodeInvalidBidPrice, err.Error())
		default:
			if h.metrics != nil {
				h.metrics.RecordBidFailure("internal_error")
			}
			api.Error(w, nethttp.StatusInternalServerError, api.CodeInternalError, "failed to create bid")
		}
		return
	}
	if h.metrics != nil {
		h.metrics.RecordBidSuccess()
	}

	h.hub.Publish(req.RoomID, ws.EventAuctionPriceUpdated, result)
	if settlement != nil {
		h.hub.Publish(req.RoomID, ws.EventAuctionSessionEnded, settlement)
		if settlement.Order != nil {
			h.hub.Publish(req.RoomID, ws.EventAuctionOrderCreated, settlement.Order)
		}
		if settlement.NextSessionID != "" {
			h.hub.Publish(req.RoomID, ws.EventAuctionSessionUpcoming, map[string]any{
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
		userID,
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
