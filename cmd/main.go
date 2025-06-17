package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/kakitomeru/gateway/internal/app"
	"github.com/kakitomeru/gateway/internal/config"
	"github.com/kakitomeru/shared/env"
	"github.com/kakitomeru/shared/logger"
)

func main() {
	logger.InitSlog("gateway", "dev", slog.LevelDebug)
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error(ctx, "failed to load gateway config", err)
		os.Exit(1)
	}

	if err := env.LoadEnv(cfg.Env); err != nil {
		logger.Error(ctx, "failed to load env", err)
		os.Exit(1)
	}

	app := app.NewApp(cfg)
	app.Start(ctx)
}
