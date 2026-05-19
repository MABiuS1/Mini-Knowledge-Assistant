package main

import (
	"log"

	"github.com/mabius/knowledge-assistant/backend/internal/config"
	httpapi "github.com/mabius/knowledge-assistant/backend/internal/http"
)

func main() {
	cfg := config.Load()
	app := httpapi.NewServer(cfg)

	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
