package service

import "errors"

var (
	ErrRoomNotFound         = errors.New("room not found")
	ErrSessionNotFound      = errors.New("session not found")
	ErrItemNotFound         = errors.New("item not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUnauthorizedToken    = errors.New("unauthorized token")
	ErrForbiddenRole        = errors.New("forbidden role")
	ErrSessionNotBidding    = errors.New("session is not bidding")
	ErrBidOwnershipMismatch = errors.New("room, session and item do not match")
	ErrDuplicateBidRequest  = errors.New("duplicate bid request")
	ErrInvalidBidPrice      = errors.New("invalid bid price")
	ErrInvalidSessionState  = errors.New("invalid session state transition")
	ErrOrderNotFound        = errors.New("order not found")
	ErrInvalidOrderState    = errors.New("invalid order state transition")
	ErrQueueExhausted       = errors.New("no next item in queue")
	ErrInvalidQueueOrder    = errors.New("invalid queue order")
)
