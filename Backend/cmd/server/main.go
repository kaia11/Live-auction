package main

import (
	"log"

	"auction-live/backend/internal/app"
	"auction-live/backend/internal/config"
	"auction-live/backend/internal/logger"
)

func main() {
	cfg := config.Load()
	logger.Info("config loaded env=%s port=%s ws_allowed_origin=%s", cfg.AppEnv, cfg.AppPort, cfg.WSAllowedOrigin)

	application := app.New(cfg)
	if err := application.Run(); err != nil {
		logger.Error("server stopped with error=%v", err)
		log.Fatal(err)
	}
}
