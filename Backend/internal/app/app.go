package app

import (
	"net/http"

	"auction-live/backend/internal/config"
	httpx "auction-live/backend/internal/http"
	"auction-live/backend/internal/http/handler"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type App struct {
	server *http.Server
}

func New(cfg config.Config) *App {
	hub := ws.NewHub()

	roomService := service.NewRoomService()
	itemService := service.NewItemService()
	sessionService := service.NewSessionService()
	bidService := service.NewBidService()
	adminService := service.NewAdminService()

	handlers := httpx.Handlers{
		Health:  handler.NewHealthHandler(cfg),
		Rooms:   handler.NewRoomHandler(roomService),
		Items:   handler.NewItemHandler(itemService),
		Session: handler.NewSessionHandler(sessionService, hub),
		Bids:    handler.NewBidHandler(bidService, hub),
		Admin:   handler.NewAdminHandler(adminService, hub),
	}

	router := httpx.NewRouter(handlers)

	return &App{
		server: &http.Server{
			Addr:    cfg.HTTPAddress(),
			Handler: router,
		},
	}
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}
