package domain

const (
	UserRoleViewer = "viewer"
	UserRoleAnchor = "anchor"
	UserRoleAdmin  = "admin"
)

const (
	RoomStatusLive    = "live"
	RoomStatusOffline = "offline"
)

const (
	QueueStateQueued    = "queued"
	QueueStateUpcoming  = "upcoming"
	QueueStateActive    = "active"
	QueueStateFinished  = "finished"
	QueueStateCancelled = "cancelled"
)

const (
	SessionStatePending     = "pending"
	SessionStateBidding     = "bidding"
	SessionStateEndedSold   = "ended_sold"
	SessionStateEndedPassed = "ended_passed"
	SessionStateCancelled   = "cancelled"
)

const (
	BidStatusAccepted = "accepted"
)

const (
	OrderStatusPendingPayment = "pending_payment"
	OrderStatusPaid           = "paid"
	OrderStatusShipped        = "shipped"
	OrderStatusCompleted      = "completed"
	OrderStatusCancelled      = "cancelled"
)
