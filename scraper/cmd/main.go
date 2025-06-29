package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-co-op/gocron"
	scraperapplication "scraper/internal/application"
	"scraper/internal/clients/github"
	"scraper/internal/clients/sender"
	"scraper/internal/clients/stackoverflow"
	scraperconfig "scraper/internal/config"
	cronModel "scraper/internal/cron"
	addlinkhandler "scraper/internal/http/handlers/add_link"
	deletehandler "scraper/internal/http/handlers/delete_chat"
	getlinkhandler "scraper/internal/http/handlers/get_links"
	newchathandler "scraper/internal/http/handlers/new_chat"
	removelinkhandler "scraper/internal/http/handlers/remove_link"
	mwlogger "scraper/internal/http/middleware/logger"
	mw "scraper/internal/http/middleware/prometheus"
	"scraper/internal/metrics"
	db "scraper/internal/storage/postgres"
	scraperUC "scraper/internal/usecase"

	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type App struct {
	ScraperServer *scraperapplication.App
}

func main() {
	cfg := scraperconfig.MustLoad()
	log := setupLogger(cfg.Env)
	ctx, cancel := context.WithCancel(context.Background())
	metricManager := metrics.NewMetricManager()

	storage, err := db.New(ctx, cfg.Scraper.AccessType, cfg.Scraper.StoragePath, cfg.Scraper.MaxConn, cfg.Scraper.MinConn)
	if err != nil {
		log.Error("Failed to initialize storage")
		return
	}

	gitClient, err := github.New(log, &cfg.Clients, metricManager)
	if err != nil {
		log.Error("Failed to initialize github client")
		return
	}

	stackClient, err := stackoverflow.New(log, &cfg.Clients, metricManager)
	if err != nil {
		log.Error("Failed to initialize stack overflow client")
		return
	}

	updateSender, err := sender.New(log, cfg)
	if err != nil {
		log.Error("Failed to initialize sender client")
		return
	}

	cron, err := setupCron(log, storage, gitClient, stackClient, updateSender, cfg.Scraper.BatchSize)
	if err != nil {
		log.Error("Failed to initialize cron")
		return
	}

	metricManager.StartCollecting()
	metricManager.CheckDBMetric(ctx, storage)

	router := setupRouter(ctx, log, storage, cfg, metricManager)

	server := scraperapplication.New(log, cfg.Scraper.Address, cfg.Scraper.Timeout, router, cron)

	app := &App{
		ScraperServer: server,
	}

	go func() {
		app.ScraperServer.MustRun()
	}()

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	cancel()
	app.ScraperServer.Stop()
	log.Info("Gracefully stopped")
}

func setupRouter(ctx context.Context, log *slog.Logger, storage db.Storage, cfg *scraperconfig.Config, manager *metrics.MetricManager) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(httprate.LimitByIP(cfg.Scraper.RateLimit, 1*time.Minute))
	router.Use(mw.PrometheusMiddleware)

	router.Route("/tg-chat", func(r chi.Router) {
		r.Post("/{id}", newchathandler.New(ctx, log, scraperUC.New(log, storage, manager)))
		r.Delete("/{id}", deletehandler.New(ctx, log, scraperUC.New(log, storage, manager)))
	})

	router.Route("/links", func(r chi.Router) {
		r.Use(httprate.LimitByIP(cfg.Scraper.LinksRateLimit, 1*time.Minute))
		r.Get("/", getlinkhandler.New(ctx, log, scraperUC.New(log, storage, manager)))
		r.Post("/", addlinkhandler.New(ctx, log, scraperUC.New(log, storage, manager)))
		r.Delete("/", removelinkhandler.New(ctx, log, scraperUC.New(log, storage, manager)))
	})

	return router
}

func setupCron(log *slog.Logger, storage db.Storage, gitClient *github.Client,
	stackClient *stackoverflow.Client, sender cronModel.Sender, batchSize uint64) (*cronModel.Cron, error) {
	baseCron := gocron.NewScheduler(time.UTC)
	cron := cronModel.New(log, baseCron, storage, gitClient, stackClient, sender, batchSize)

	_, err := cron.Cron.Every(1).Minutes().Do(cron.UpdateCron)
	if err != nil {
		return nil, err
	}

	return cron, nil
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
