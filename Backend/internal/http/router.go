package http

import (
	"net/http"

	"auction-live/backend/internal/http/handler"
)

type Handlers struct {
	Health  *handler.HealthHandler
	Rooms   *handler.RoomHandler
	Items   *handler.ItemHandler
	Session *handler.SessionHandler
	Bids    *handler.BidHandler
	Admin   *handler.AdminHandler
}

func NewRouter(handlers Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health.GetHealth)

	mux.HandleFunc("GET /rooms", handlers.Rooms.ListRooms)
	mux.HandleFunc("GET /rooms/{roomId}", handlers.Rooms.GetRoomDetail)
	mux.HandleFunc("GET /rooms/{roomId}/items", handlers.Items.ListRoomItems)
	mux.HandleFunc("GET /rooms/{roomId}/items/{itemId}", handlers.Items.GetItemDetail)
	mux.HandleFunc("GET /rooms/{roomId}/current-session", handlers.Session.GetCurrentSession)
	mux.HandleFunc("GET /rooms/{roomId}/events", handlers.Session.GetRoomEvents)
	mux.HandleFunc("GET /sessions/{sessionId}/ranking", handlers.Session.GetRanking)
	mux.HandleFunc("GET /sessions/{sessionId}/my-status", handlers.Session.GetMyStatus)
	mux.HandleFunc("GET /users/me/bids", handlers.Bids.ListMyBids)
	mux.HandleFunc("POST /bids", handlers.Bids.CreateBid)

	mux.HandleFunc("POST /admin/rooms/{roomId}/items", handlers.Admin.CreateItem)
	mux.HandleFunc("PATCH /admin/items/{itemId}", handlers.Admin.UpdateItem)
	mux.HandleFunc("POST /admin/rooms/{roomId}/queue/reorder", handlers.Admin.ReorderQueue)
	mux.HandleFunc("POST /admin/rooms/{roomId}/queue/next", handlers.Admin.ActivateNextItem)
	mux.HandleFunc("POST /admin/sessions/{sessionId}/start", handlers.Admin.StartSession)
	mux.HandleFunc("POST /admin/sessions/{sessionId}/cancel", handlers.Admin.CancelSession)
	mux.HandleFunc("GET /admin/rooms/{roomId}/sessions", handlers.Admin.ListRoomSessions)
	mux.HandleFunc("GET /admin/orders", handlers.Admin.ListOrders)
	mux.HandleFunc("GET /admin/stats/overview", handlers.Admin.GetStatsOverview)
	mux.HandleFunc("GET /admin/stats/timeline", handlers.Admin.GetStatsTimeline)

	return mux
}
