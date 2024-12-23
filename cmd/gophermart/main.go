package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/GlebRadaev/gofermart/internal/app"
	"github.com/rs/zerolog/log"
	"go.uber.org/zap"
)

//	@title			GoFemart API
//	@version		1.0
//	@description	API Server

// @host		localhost:8080
// @BasePath	/
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := app.New()
	err := app.Start(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Can't start application")
		zap.L().Fatal("Can't start application: ", zap.Error(err))
	}

	err = app.Wait(ctx, cancel)
	if err != nil {
		zap.L().Fatal("All systems closed with errors. LastError:", zap.Error(err))
	}

	zap.L().Info("All systems closed without errors")
}
