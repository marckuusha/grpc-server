package grpcapp

import (
	"fmt"
	server "grpc-server/internal/grpc"
	"grpc-server/internal/logic"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	log        *zap.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *zap.Logger, port int, logicapp *logic.LogicApp) *App {

	gRPCServer := grpc.NewServer()

	server.Register(gRPCServer, logicapp)

	return &App{
		port:       port,
		log:        log,
		gRPCServer: gRPCServer,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {

	const op = "grpcApp.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s : %w", op, err)
	}

	a.log.Sugar().Infof("grpc server start in: addr = %s", l.Addr().String())

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s : %w ", op, err)
	}

	return nil
}

func (a *App) Stop() {
	a.log.Info("stop grpc server")
	a.gRPCServer.GracefulStop()
}
