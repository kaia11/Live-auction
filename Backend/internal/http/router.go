package http

import (
	"net/http"

	"auction-live/backend/internal/http/handler"
)

type Handlers struct {
	Auth      *handler.AuthHandler
	Health    *handler.HealthHandler
	Metrics   *handler.MetricsHandler
	Rooms     *handler.RoomHandler
	Items     *handler.ItemHandler
	Orders    *handler.OrderHandler
	Session   *handler.SessionHandler
	Bids      *handler.BidHandler
	Admin     *handler.AdminHandler
	Upload    *handler.UploadHandler
	WebSocket *handler.WebSocketHandler
}

func NewRouter(handlers Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /auth/login", handlers.Auth.Login)
	mux.HandleFunc("POST /auth/register", handlers.Auth.Register)
	mux.HandleFunc("GET /users/me", handlers.Auth.GetCurrentUser)
	mux.HandleFunc("GET /health", handlers.Health.GetHealth)
	mux.HandleFunc("GET /metrics", handlers.Metrics.GetMetrics)
	mux.HandleFunc("GET /uploads/{file...}", handlers.Upload.ServeUpload)
	mux.HandleFunc("GET /rooms", handlers.Rooms.ListRooms)
	mux.HandleFunc("GET /admin/my/rooms", handlers.Rooms.ListMyRooms)
	mux.HandleFunc("GET /rooms/{roomId}", handlers.Rooms.GetRoomDetail)
	mux.HandleFunc("GET /rooms/{roomId}/live-snapshot", handlers.Rooms.GetLiveSnapshot)
	mux.HandleFunc("GET /rooms/{roomId}/items", handlers.Items.ListRoomItems)
	mux.HandleFunc("GET /rooms/{roomId}/items/{itemId}", handlers.Items.GetItemDetail)
	mux.HandleFunc("GET /rooms/{roomId}/current-session", handlers.Session.GetCurrentSession)
	mux.HandleFunc("GET /rooms/{roomId}/events", handlers.Session.GetRoomEvents)
	mux.HandleFunc("POST /rooms/{roomId}/comments", handlers.Session.CreateRoomComment)
	mux.HandleFunc("GET /ws", handlers.WebSocket.ServeRoomStream)
	mux.HandleFunc("GET /sessions/{sessionId}/ranking", handlers.Session.GetRanking)
	mux.HandleFunc("GET /sessions/{sessionId}/my-status", handlers.Session.GetMyStatus)
	mux.HandleFunc("GET /users/me/bids", handlers.Bids.ListMyBids)
	mux.HandleFunc("GET /users/me/orders", handlers.Orders.ListMyOrders)
	mux.HandleFunc("POST /bids", handlers.Bids.CreateBid)
	mux.HandleFunc("POST /sessions/{sessionId}/auto-proxy", handlers.Bids.ConfigureAutoProxy)

	mux.HandleFunc("POST /admin/uploads/image", handlers.Upload.UploadImage)
	mux.HandleFunc("POST /admin/rooms/{roomId}/items", handlers.Admin.CreateItem)
	mux.HandleFunc("PATCH /admin/items/{itemId}", handlers.Admin.UpdateItem)
	mux.HandleFunc("POST /admin/rooms/{roomId}/queue/reorder", handlers.Admin.ReorderQueue)
	mux.HandleFunc("POST /admin/rooms/{roomId}/queue/next", handlers.Admin.ActivateNextItem)
	mux.HandleFunc("POST /admin/rooms/{roomId}/start", handlers.Admin.StartRoomLive)
	mux.HandleFunc("POST /admin/rooms/{roomId}/stop", handlers.Admin.StopRoomLive)
	mux.HandleFunc("POST /admin/sessions/{sessionId}/start", handlers.Admin.StartSession)
	mux.HandleFunc("POST /admin/sessions/{sessionId}/cancel", handlers.Admin.CancelSession)
	mux.HandleFunc("POST /admin/sessions/{sessionId}/settle", handlers.Admin.SettleSession)
	mux.HandleFunc("GET /admin/rooms/{roomId}/sessions", handlers.Admin.ListRoomSessions)
	mux.HandleFunc("GET /admin/orders", handlers.Admin.ListOrders)
	mux.HandleFunc("POST /admin/orders/{orderId}/status", handlers.Orders.UpdateOrderStatus)
	mux.HandleFunc("GET /admin/stats/overview", handlers.Admin.GetStatsOverview)
	mux.HandleFunc("GET /admin/stats/timeline", handlers.Admin.GetStatsTimeline)

	return mux
}
