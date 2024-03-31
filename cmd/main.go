package main

import (
	"grpc-server/internal/app"
	"grpc-server/internal/config"

	"go.uber.org/zap"
)

func main() {

	// load config
	cfg := config.MustLoad()

	// init logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// init app
	appl := app.NewApp(logger, cfg)

	appl.StartServer()
}
