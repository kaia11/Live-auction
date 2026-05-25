package main

import (
	"log"

	"auction-live/backend/internal/app"
	"auction-live/backend/internal/config"
)

func main() {
	cfg := config.Load()

	application := app.New(cfg)
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
