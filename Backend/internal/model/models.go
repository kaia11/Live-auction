package model

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

type LiveRoom struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	CoverImage       string `json:"coverImage"`
	VideoURL         string `json:"videoUrl"`
	Status           string `json:"status"`
	AnchorUserID     string `json:"anchorUserId"`
	AnchorName       string `json:"anchorName"`
	OnlineCount      int    `json:"onlineCount"`
	Thumbnail        string `json:"thumbnail,omitempty"`
	CurrentSessionID string `json:"currentSessionId"`
}

type AuctionItem struct {
	ID                      string `json:"id"`
	RoomID                  string `json:"roomId"`
	Title                   string `json:"title"`
	CoverImage              string `json:"coverImage"`
	Description             string `json:"description"`
	StartPrice              int64  `json:"startPrice"`
	IncrementStep           int64  `json:"incrementStep"`
	CeilingPrice            *int64 `json:"ceilingPrice,omitempty"`
	DurationSeconds         int    `json:"durationSeconds"`
	ExtensionSeconds        int    `json:"extensionSeconds"`
	ExtensionTriggerSeconds int    `json:"extensionTriggerSeconds"`
	QueueStatus             string `json:"queueStatus"`
}

type AuctionSession struct {
	ID                string `json:"id"`
	RoomID            string `json:"roomId"`
	ItemID            string `json:"itemId"`
	Status            string `json:"status"`
	CurrentPrice      int64  `json:"currentPrice"`
	LeaderUserID      string `json:"leaderUserId"`
	StartTime         string `json:"startTime"`
	EndTime           string `json:"endTime"`
	ParticipantCount  int    `json:"participantCount"`
	ViewerCount       int    `json:"viewerCount"`
	IncrementStep     int64  `json:"incrementStep"`
	ExtensionSeconds  int    `json:"extensionSeconds"`
	ExtensionTrigger  int    `json:"extensionTriggerSeconds"`
	CeilingPrice      *int64 `json:"ceilingPrice,omitempty"`
	SupportsAutoProxy bool   `json:"supportsAutoProxy"`
}

type Bid struct {
	ID         string `json:"id"`
	SessionID  string `json:"sessionId"`
	RoomID     string `json:"roomId"`
	ItemID     string `json:"itemId"`
	UserID     string `json:"userId"`
	BidPrice   int64  `json:"bidPrice"`
	RequestID  string `json:"requestId"`
	RankAfter  int    `json:"rankAfter"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
}

type UserBidHistory struct {
	ID        string `json:"id"`
	ItemID    string `json:"itemId"`
	ItemTitle string `json:"itemTitle"`
	ItemImage string `json:"itemImage"`
	BidPrice  int64  `json:"bidPrice"`
	Result    string `json:"result"`
	BidTime   string `json:"bidTime"`
}

type BidResult struct {
	RoomID            string `json:"roomId"`
	SessionID         string `json:"sessionId"`
	ItemID            string `json:"itemId"`
	UserID            string `json:"userId"`
	AcceptedBidPrice  int64  `json:"acceptedBidPrice"`
	RequestID         string `json:"requestId"`
	CurrentPrice      int64  `json:"currentPrice"`
	IsLeading         bool   `json:"isLeading"`
	ExtensionApplied  bool   `json:"extensionApplied"`
	CeilingReached    bool   `json:"ceilingReached"`
	NextMinimumBid    int64  `json:"nextMinimumBid"`
	VibrateSignalHint string `json:"vibrateSignalHint"`
}

type RankingEntry struct {
	UserID     string `json:"userId"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Rank       int    `json:"rank"`
	HighestBid int64  `json:"highestBid"`
}

type SessionUserStatus struct {
	SessionID         string `json:"sessionId"`
	UserID            string `json:"userId"`
	MyHighestBid      int64  `json:"myHighestBid"`
	MyRank            int    `json:"myRank"`
	IsLeading         bool   `json:"isLeading"`
	CurrentPrice      int64  `json:"currentPrice"`
	NextMinimumBid    int64  `json:"nextMinimumBid"`
	VibrateSignalHint string `json:"vibrateSignalHint"`
	AutoProxyEnabled  bool   `json:"autoProxyEnabled"`
	AutoProxyMaxPrice int64  `json:"autoProxyMaxPrice"`
}

type AutoProxyConfig struct {
	SessionID string `json:"sessionId"`
	RoomID    string `json:"roomId"`
	ItemID    string `json:"itemId"`
	UserID    string `json:"userId"`
	MaxPrice  int64  `json:"maxPrice"`
	EnabledAt string `json:"enabledAt"`
}

type AuctionResult struct {
	SessionID        string `json:"sessionId"`
	ItemID           string `json:"itemId"`
	ResultStatus     string `json:"resultStatus"`
	WinnerUserID     string `json:"winnerUserId"`
	FinalPrice       int64  `json:"finalPrice"`
	ParticipantCount int    `json:"participantCount"`
}

type AuctionOrder struct {
	ID          string `json:"id"`
	SessionID   string `json:"sessionId"`
	RoomID      string `json:"roomId"`
	ItemID      string `json:"itemId"`
	BuyerUserID string `json:"buyerUserId"`
	Amount      int64  `json:"amount"`
	Status      string `json:"status"`
	CreateTime  string `json:"createTime"`
}

type RoomComment struct {
	ID         int64  `json:"id"`
	RoomID     string `json:"roomId"`
	UserID     string `json:"userId"`
	Nickname   string `json:"nickname"`
	Content    string `json:"content"`
	CreateTime string `json:"createTime"`
}

type OperationLog struct {
	ID         int64  `json:"id"`
	Module     string `json:"module"`
	Action     string `json:"action"`
	OperatorID string `json:"operatorId"`
	RoomID     string `json:"roomId"`
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Content    string `json:"content"`
	CreateTime string `json:"createTime"`
}
