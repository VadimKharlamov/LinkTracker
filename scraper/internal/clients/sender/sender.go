package sender

import (
	"context"
	"fmt"
	"log/slog"
	"scraper/internal/config"

	"scraper/internal/model/scraper"
)

type Sender interface {
	Updates(ctx context.Context, req *scraper.LinkUpdate, isFailed bool) error
}

const (
	messageTransportHTTP  = "HTTP"
	messageTransportKafka = "KAFKA"
)

type FallbackSender struct {
	Primary  Sender
	Fallback Sender
	Logger   *slog.Logger
}

func New(logger *slog.Logger, cfg *config.Config) (Sender, error) {
	var (
		httpSender  Sender
		kafkaSender Sender
	)

	httpSender, httpErr := NewClient(logger, &cfg.Clients)
	kafkaSender, kafkaErr := NewProducer([]string{cfg.Clients.Kafka.Address}, cfg.Clients.Kafka.Topic,
		cfg.Clients.Kafka.DLQTopic, cfg.Clients.Kafka.Timeout)

	switch cfg.Scraper.TransportType {
	case messageTransportHTTP:
		if httpErr != nil {
			return nil, fmt.Errorf("failed to init HTTP sender: %w", httpErr)
		}

		if kafkaErr != nil {
			logger.Warn("Kafka fallback unavailable", slog.String("error", kafkaErr.Error()))
		}

		return &FallbackSender{Primary: httpSender, Fallback: kafkaSender, Logger: logger}, nil

	case messageTransportKafka:
		if kafkaErr != nil {
			return nil, fmt.Errorf("failed to init Kafka sender: %w", kafkaErr)
		}

		if httpErr != nil {
			logger.Warn("HTTP fallback unavailable", slog.String("error", httpErr.Error()))
		}

		return &FallbackSender{Primary: kafkaSender, Fallback: httpSender, Logger: logger}, nil

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", cfg.Scraper.TransportType)
	}
}

func (f *FallbackSender) Updates(ctx context.Context, req *scraper.LinkUpdate, isFailed bool) error {
	err := f.Primary.Updates(ctx, req, isFailed)
	if err == nil {
		return nil
	}

	f.Logger.Warn("primary sender failed, falling back", slog.String("error", err.Error()))

	fallbackErr := f.Fallback.Updates(ctx, req, isFailed)
	if fallbackErr != nil {
		return fmt.Errorf("primary failed: %v, fallback also failed: %w", err, fallbackErr)
	}

	return nil
}
