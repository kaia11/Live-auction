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
	ws.SeedDemoMessages(hub)

	userService := service.NewUserService()
	roomService := service.NewRoomService()
	itemService := service.NewItemService()
	sessionService := service.NewSessionService()
	bidService := service.NewBidService()
	adminService := service.NewAdminService()

	handlers := httpx.Handlers{
		Auth:    handler.NewAuthHandler(userService),
		Health:  handler.NewHealthHandler(cfg),
		Rooms:   handler.NewRoomHandler(roomService),
		Items:   handler.NewItemHandler(itemService),
		Session: handler.NewSessionHandler(sessionService, userService, hub),
		Bids:    handler.NewBidHandler(bidService, userService, hub),
		Admin:   handler.NewAdminHandler(adminService, hub),
	}

	router := httpx.NewRouter(handlers)
	handlerWithCORS := httpx.WithCORS(router, cfg.WSAllowedOrigin)

	return &App{
		server: &http.Server{
			Addr:    cfg.HTTPAddress(),
			Handler: handlerWithCORS,
		},
	}
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}
