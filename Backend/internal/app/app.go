package app

import (
	"net/http"
	"time"

	"auction-live/backend/internal/config"
	httpx "auction-live/backend/internal/http"
	"auction-live/backend/internal/http/handler"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type App struct {
	server    *http.Server
	scheduler *service.SettlementScheduler
}

func New(cfg config.Config) *App {
	hub := ws.NewHub()
	ws.SeedDemoMessages(hub)
	logger.Info("seeded demo room messages room_id=%s count=%d", "room-001", 3)

	userService := service.NewUserService()
	roomService := service.NewRoomService()
	liveSnapshotService := service.NewLiveSnapshotService(hub)
	itemService := service.NewItemService()
	sessionService := service.NewSessionService()
	bidService := service.NewBidService()
	adminService := service.NewAdminService()
	orderService := service.NewOrderService()
	scheduler := service.NewSettlementScheduler(service.SharedStore(), hub, time.Second)

	handlers := httpx.Handlers{
		Auth:    handler.NewAuthHandler(userService),
		Health:  handler.NewHealthHandler(cfg),
		Rooms:   handler.NewRoomHandler(roomService, liveSnapshotService, userService),
		Items:   handler.NewItemHandler(itemService),
		Orders:  handler.NewOrderHandler(orderService, userService, hub),
		Session: handler.NewSessionHandler(sessionService, userService, hub),
		Bids:    handler.NewBidHandler(bidService, userService, hub),
		Admin:   handler.NewAdminHandler(adminService, hub),
	}

	router := httpx.NewRouter(handlers)
	handlerWithCORS := httpx.WithCORS(router, cfg.WSAllowedOrigin)
	handlerWithLogging := httpx.WithRequestLogging(handlerWithCORS)

	return &App{
		server: &http.Server{
			Addr:    cfg.HTTPAddress(),
			Handler: handlerWithLogging,
		},
		scheduler: scheduler,
	}
}

func (a *App) Run() error {
	if a.scheduler != nil {
		a.scheduler.Start()
		defer a.scheduler.Stop()
	}
	logger.Info("http server starting addr=%s", a.server.Addr)
	return a.server.ListenAndServe()
}
