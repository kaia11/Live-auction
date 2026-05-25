package service

import "errors"

var (
	ErrRoomNotFound         = errors.New("room not found")
	ErrSessionNotFound      = errors.New("session not found")
	ErrItemNotFound         = errors.New("item not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrSessionNotBidding    = errors.New("session is not bidding")
	ErrBidOwnershipMismatch = errors.New("room, session and item do not match")
	ErrDuplicateBidRequest  = errors.New("duplicate bid request")
	ErrInvalidBidPrice      = errors.New("invalid bid price")
	ErrInvalidSessionState  = errors.New("invalid session state transition")
	ErrQueueExhausted       = errors.New("no next item in queue")
	ErrInvalidQueueOrder    = errors.New("invalid queue order")
)
