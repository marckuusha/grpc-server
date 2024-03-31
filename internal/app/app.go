package app

import (
	"context"
	grpcapp "grpc-server/internal/app/grpc"
	"grpc-server/internal/config"
	"grpc-server/internal/logic"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.uber.org/zap"
)

type App struct {
	GRPC     *grpcapp.App
	logicApp *logic.LogicApp

	cfg *config.Config
	log *zap.Logger
}

func NewApp(log *zap.Logger, cfg *config.Config) *App {
	logicapp := logic.NewLogic()

	app := grpcapp.New(log, cfg.GRPCServer.Port, logicapp)

	return &App{
		GRPC:     app,
		cfg:      cfg,
		log:      log,
		logicApp: logicapp,
	}
}

func (a *App) StartServer() {
	serv := &http.Server{
		Addr:         a.cfg.Address,
		Handler:      a.setupRouter(),
		ReadTimeout:  a.cfg.HTTPServer.Timeout,
		WriteTimeout: a.cfg.HTTPServer.Timeout,
		IdleTimeout:  a.cfg.HTTPServer.IdleTimeout,
	}

	listErr := make(chan error, 1)
	go func() {
		a.log.Sugar().Infof("http сервер запущен: %s", a.cfg.Address)
		listErr <- serv.ListenAndServe()
	}()

	grpcErr := make(chan error, 1)
	go func() {
		a.log.Sugar().Infof("grpc сервер запущен на порту: %d", a.cfg.GRPCServer.Port)
		grpcErr <- a.GRPC.Run()
	}()

	const timeout = time.Second * 15
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	select {
	case <-stop:
		a.GRPC.Stop()

		serv.SetKeepAlivesEnabled(false)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := serv.Shutdown(ctx); err != nil {
			a.log.Sugar().Errorf("ошибка во время закрытия сервиса: %s", err)
			return
		}
	case <-listErr:
		a.GRPC.Stop()
		return
	case <-grpcErr:
		serv.SetKeepAlivesEnabled(false)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := serv.Shutdown(ctx); err != nil {
			a.log.Sugar().Errorf("ошибка во время закрытия сервиса: %s", err)
			return
		}
	}
}

func (a *App) setupRouter() *chi.Mux {
	// init router
	router := chi.NewRouter()
	// handles
	router.Get("/test", a.HandleGenerateText)
	return router
}

type Response struct {
	Status   int    `json:"status"`
	Error    string `json:"error,omitempty"`
	Template string `json:"template,omitempty"`
}

func (a *App) HandleGenerateText(w http.ResponseWriter, r *http.Request) {

	queryParams := r.URL.Query()
	name := queryParams.Get("name")

	if name == "" {
		render.JSON(w, r, Response{
			Status: http.StatusBadRequest,
			Error:  "name is empty",
		})
		return
	}

	templ, err := a.logicApp.GenerateText(name)
	if err != nil {
		render.JSON(w, r, Response{
			Status: http.StatusInternalServerError,
			Error:  "cannot generate template",
		})
		return
	}

	render.JSON(w, r, Response{
		Status:   http.StatusOK,
		Template: templ,
	})
}
