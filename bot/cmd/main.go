package main

import (
	botapplication "bot/internal/application"
	"bot/internal/clients/kafka"
	"bot/internal/clients/scraper"
	botconfig "bot/internal/config"
	updateHandler "bot/internal/http/handlers/update"
	"bot/internal/http/middleware/logger"
	"bot/internal/metrics"
	botModel "bot/internal/model/bot"
	db "bot/internal/storage/redis"
	bothandlers "bot/internal/tg/handlers"
	botUC "bot/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"gopkg.in/telebot.v3"
	"sync"

	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal   = "local"
	envDev     = "dev"
	envProd    = "prod"
	KafkaGroup = "1"
)

type App struct {
	BotServer *botapplication.App
}

var states = make(map[int64]*botModel.UserState)

func main() {
	cfg := botconfig.MustLoad()
	metricManager := metrics.NewMetricManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := setupLogger(cfg.Env)

	if log == nil {
		fmt.Println("Failed to setup logger")
		return
	}

	storage := db.New(cfg.Bot.StoragePath, cfg.Bot.MaxIdle, cfg.Bot.MaxActive)

	tgBot, err := createTgBot(cfg)
	if err != nil {
		log.Error("Failed to create bot")
		return
	}

	scraperClient, err := scraperclient.New(log, &cfg.Clients)
	if err != nil {
		log.Error("Failed to create scraper client")
		return
	}

	bot := bothandlers.Bot{
		Handler:       tgBot,
		Logger:        log,
		States:        states,
		MetricManager: metricManager,
	}

	router := setupRouter(ctx, log, &bot, scraperClient, storage, cfg)
	errChan := make(chan error, 2)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		if err = kafka.RunConsumerGroup(ctx, log, []string{cfg.Clients.Kafka.Address}, KafkaGroup,
			[]string{cfg.Clients.Kafka.Topic}, botUC.New(log, &bot, scraperClient, storage).Update); err != nil {
			errChan <- fmt.Errorf("base consumer error: %w", err)
		}
	}()

	go func() {
		if err = kafka.RunConsumerGroup(ctx, log, []string{cfg.Clients.Kafka.Address}, KafkaGroup,
			[]string{cfg.Clients.Kafka.DLQTopic}, botUC.New(log, &bot, scraperClient, storage).ProcessFail); err != nil {
			errChan <- fmt.Errorf("base consumer error: %w", err)
		}
	}()

	select {
	case err = <-errChan:
		log.Error("Kafka consumer failed to start", err)
		cancel()
		return
	case <-time.After(2 * time.Second):
	}

	server := botapplication.New(log, cfg.Bot.Address, cfg.Bot.Timeout, router, &bot)

	app := &App{
		BotServer: server,
	}

	go app.BotServer.MustRun()

	// Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	cancel()
	app.BotServer.Stop()
	log.Info("Gracefully stopped")
}

func createTgBot(cfg *botconfig.Config) (*telebot.Bot, error) {
	pref := telebot.Settings{
		Token:  cfg.Bot.Token,
		Poller: &telebot.LongPoller{Timeout: cfg.Bot.Timeout},
	}

	tgBot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, err
	}

	return tgBot, nil
}

func setupRouter(ctx context.Context, log *slog.Logger, bot *bothandlers.Bot,
	client *scraperclient.Client, storage *db.Storage, cfg *botconfig.Config) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(httprate.LimitByIP(cfg.Bot.RateLimit, 1*time.Minute))

	router.Route("/updates", func(r chi.Router) {
		r.Post("/", updateHandler.New(log, botUC.New(log, bot, client, storage)))
	})

	bot.Handler.Handle("/start", bot.StartHandler(ctx, botUC.New(log, bot, client, storage)))
	bot.Handler.Handle("/help", bot.HelpHandler)
	bot.Handler.Handle("/track", bot.TrackHandler)
	bot.Handler.Handle(telebot.OnText, bot.StatesHandler(ctx, botUC.New(log, bot, client, storage)))
	bot.Handler.Handle("/untrack", bot.UntrackHandler(ctx, botUC.New(log, bot, client, storage)))
	bot.Handler.Handle("/list", bot.ListHandler(ctx, botUC.New(log, bot, client, storage)))

	return router
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
