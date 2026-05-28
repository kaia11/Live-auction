package ws

type CommentPayload struct {
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
	Content  string `json:"content"`
}

func SeedDemoMessages(hub *Hub) {
	if hub == nil {
		return
	}

	demoComments := []CommentPayload{
		{UserID: "user-001", Nickname: "阿宁", Content: "这个吊坠水头真不错，想再看看细节"},
		{UserID: "user-002", Nickname: "小满", Content: "主播把背面也转一下，准备出价了"},
		{UserID: "user-003", Nickname: "阿青", Content: "850 还有人跟吗？这块料子挺值的"},
	}

	for _, comment := range demoComments {
		hub.Publish("room-001", EventRoomCommentReceived, comment)
	}
}
