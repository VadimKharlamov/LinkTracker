package application

import (
	bothandlers "bot/internal/tg/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type App struct {
	log    *slog.Logger
	server *http.Server
	bot    *bothandlers.Bot
}

func New(
	log *slog.Logger,
	address string,
	timeout time.Duration,
	router *chi.Mux,
	bot *bothandlers.Bot,
) *App {
	srv := &http.Server{
		Addr:         address,
		Handler:      router,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout,
	}

	return &App{
		log:    log,
		server: srv,
		bot:    bot,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "server.Run"

	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()

		if err := a.server.ListenAndServe(); err != nil {
			a.log.Error("failed to start server")
		}
	}()

	wg.Add(1)

	metricsMux := http.NewServeMux()

	metricsMux.Handle("/metrics", promhttp.Handler())

	go func() {
		defer wg.Done()

		a.log.Info("metrics listening on :9200")

		if err := http.ListenAndServe(":9200", metricsMux); err != nil {
			a.log.Error("metrics server error", slog.String("err", err.Error()))
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		a.bot.Handler.Start()
	}()

	a.log.With(slog.String("op", op)).
		Info("bot and server started", slog.String("addr", a.server.Addr))

	go func() {
		wg.Wait()
		a.log.With(slog.String("op", op)).
			Info("bot and server stopped")
	}()

	return nil
}

func (a *App) Stop() {
	const op = "server.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping bot server", slog.String("addr", a.server.Addr))

	err := a.server.Shutdown(context.Background())
	if err != nil {
		a.log.Info("failed to stop server", slog.String("op", op), slog.String("err", err.Error()))
	}

	a.bot.Handler.Stop()
}
