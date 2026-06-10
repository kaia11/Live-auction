package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"auction-live/backend/internal/config"
	httpx "auction-live/backend/internal/http"
	"auction-live/backend/internal/http/handler"
	"auction-live/backend/internal/logger"
	"auction-live/backend/internal/monitoring"
	"auction-live/backend/internal/persistence"
	"auction-live/backend/internal/realtime"
	"auction-live/backend/internal/repository"
	"auction-live/backend/internal/service"
	"auction-live/backend/internal/ws"
)

type App struct {
	server    *http.Server
	scheduler *service.SettlementScheduler
}

func New(cfg config.Config) *App {
	redisClient := realtime.NewClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	runtime := realtime.NewRuntime(redisClient)
	var pingErr error
	for attempt := 1; attempt <= 10; attempt++ {
		pingErr = runtime.Ping()
		if pingErr == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	if pingErr != nil {
		panic(fmt.Errorf("redis ping failed: %w", pingErr))
	}

	hub := ws.NewHub(runtime)
	ws.SeedDemoMessages(hub)
	logger.Info("seeded demo room messages room_id=%s count=%d", "room-001", 3)
	metrics := monitoring.NewMetrics()

	tokenService := service.NewTokenService(cfg.JWTSecret)
	store := service.SharedStore()
	userRepo := repository.UserRepository(repository.NewMemoryUserRepository(store))
	roomRepo := repository.RoomRepository(repository.NewMemoryRoomRepository(store))
	itemRepo := repository.ItemRepository(repository.NewMemoryItemRepository(store))
	sessionRepo := repository.SessionRepository(repository.NewMemorySessionRepository(store))
	orderRepo := repository.OrderRepository(repository.NewMemoryOrderRepository(store))
	bidRepo := repository.BidRepository(repository.NewMemoryBidRepository(store))
	resultRepo := repository.ResultRepository(repository.NewMemoryResultRepository())
	commentRepo := repository.CommentRepository(repository.NewMemoryCommentRepository())
	logRepo := repository.OperationLogRepository(repository.NewMemoryOperationLogRepository())
	if db, err := persistence.OpenMySQL(cfg.MySQLDSN); err == nil {
		userRepo = repository.NewMySQLUserRepository(db)
		roomRepo = repository.NewMySQLRoomRepository(db)
		itemRepo = repository.NewMySQLItemRepository(db)
		sessionRepo = repository.NewMySQLSessionRepository(db)
		orderRepo = repository.NewMySQLOrderRepository(db)
		bidRepo = repository.NewMySQLBidRepository(db)
		resultRepo = repository.NewMySQLResultRepository(db)
		commentRepo = repository.NewMySQLCommentRepository(db)
		logRepo = repository.NewMySQLOperationLogRepository(db)
	} else if cfg.RequirePersistentLedger {
		panic(fmt.Errorf("mysql required for user accounts: %w", err))
	} else {
		logger.Error("mysql init skipped, fallback to memory mode error=%v", err)
	}

	if err := service.SyncMemoryBootstrap(store, userRepo, roomRepo, itemRepo, sessionRepo, bidRepo); err != nil {
		panic(fmt.Errorf("memory bootstrap failed: %w", err))
	}

	userService := service.NewUserService(tokenService, userRepo, roomRepo, redisClient)
	if err := userService.WarmUserCacheAndMigratePasswords(); err != nil {
		panic(fmt.Errorf("user bootstrap failed: %w", err))
	}
	roomService := service.NewRoomService(roomRepo)
	liveSnapshotService := service.NewLiveSnapshotService(hub, runtime, roomRepo, itemRepo, sessionRepo, bidRepo, userRepo)
	itemService := service.NewItemService(itemRepo)
	sessionService := service.NewSessionService(runtime, sessionRepo, bidRepo, userRepo)
	commentService := service.NewCommentService(commentRepo)
	auditService := service.NewAuditService(logRepo)
	bidService := service.NewBidService(runtime, cfg.RepeatBid, bidRepo, roomRepo, itemRepo, userRepo, sessionRepo, resultRepo, orderRepo)
	adminService := service.NewAdminService(runtime, roomRepo, itemRepo, sessionRepo, resultRepo, orderRepo)
	orderService := service.NewOrderService(orderRepo)
	scheduler := service.NewSettlementScheduler(store, hub, runtime, roomRepo, itemRepo, sessionRepo, resultRepo, orderRepo, metrics, time.Second)
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		panic(fmt.Errorf("upload dir init failed: %w", err))
	}
	if err := service.SyncRealtimeBootstrap(runtime, roomRepo, itemRepo, sessionRepo, bidRepo, userRepo, store); err != nil {
		panic(fmt.Errorf("realtime bootstrap failed: %w", err))
	}

	handlers := httpx.Handlers{
		Auth:      handler.NewAuthHandler(userService),
		Health:    handler.NewHealthHandler(cfg),
		Metrics:   handler.NewMetricsHandler(metrics),
		Rooms:     handler.NewRoomHandler(roomService, liveSnapshotService, userService),
		Items:     handler.NewItemHandler(itemService),
		Orders:    handler.NewOrderHandler(orderService, adminService, auditService, userService, hub),
		Session:   handler.NewSessionHandler(sessionService, commentService, userService, hub),
		Bids:      handler.NewBidHandler(bidService, userService, hub, metrics),
		Admin:     handler.NewAdminHandler(adminService, auditService, userService, hub),
		Upload:    handler.NewUploadHandler(cfg.UploadDir, userService),
		WebSocket: handler.NewWebSocketHandler(userService, roomService, hub, metrics),
	}

	router := httpx.NewRouter(handlers)
	handlerWithCORS := httpx.WithCORS(router, cfg.WSAllowedOrigin)
	handlerWithLogging := httpx.WithRequestLogging(handlerWithCORS, metrics)

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
