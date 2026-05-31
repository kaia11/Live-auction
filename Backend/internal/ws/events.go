package ws

const (
	// 房间评论新增后广播给当前直播间。
	EventRoomCommentReceived     = "room_comment_received"
	// 拍品队列状态变化后广播，例如新建拍品、队列重排。
	EventRoomItemQueueUpdated    = "room_item_queue_updated"
	// 出价成功后广播最新价格、领先者和下一口价。
	EventAuctionPriceUpdated     = "auction_price_updated"
	// 当前场次结束后，如果已经切到下一件待开拍，则广播“下一场即将开始”信息。
	EventAuctionSessionUpcoming  = "auction_session_upcoming"
	// 主播/管理端真正启动场次后广播。当前代码统一使用 activated 命名。
	EventAuctionSessionActivated = "auction_session_activated"
	// 场次结束后广播结算结果。
	EventAuctionSessionEnded     = "auction_session_ended"
	// 场次成交并创建订单后广播新订单。
	EventAuctionOrderCreated     = "auction_order_created"
	// 订单状态变化后广播更新后的订单。
	EventAuctionOrderUpdated     = "auction_order_updated"
)
