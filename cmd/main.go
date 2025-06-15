package main

import (
	"context"
	"log"

	"github.com/kakitomeru/gateway/internal/app"
	"github.com/kakitomeru/gateway/internal/config"
	"github.com/kakitomeru/shared/env"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load gateway config: %v", err)
	}

	if err := env.LoadEnv(cfg.Env); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	app := app.NewApp(cfg)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Fatalf("failed to start app: %v", err)
	}
}
