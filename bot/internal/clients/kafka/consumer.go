package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"bot/internal/model/bot"
	"github.com/Shopify/sarama"
)

type linkUpdateHandler struct {
	process func(upd *bot.LinkUpdate) error
}

func (h *linkUpdateHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *linkUpdateHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *linkUpdateHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var upd bot.LinkUpdate

		if err := json.Unmarshal(msg.Value, &upd); err != nil {
			sess.MarkMessage(msg, "unmarshal error")

			continue
		}

		if err := h.process(&upd); err != nil {
			continue
		}

		sess.MarkMessage(msg, "")
	}

	return nil
}

func RunConsumerGroup(ctx context.Context, log *slog.Logger, brokers []string, groupID string, topics []string,
	processFn func(upd *bot.LinkUpdate) error,
) error {
	cfg := sarama.NewConfig()

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		log.Error("cannot create consumer group")
		return fmt.Errorf("cannot create consumer group: %w", err)
	}
	defer consumerGroup.Close()

	handler := &linkUpdateHandler{process: processFn}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err = consumerGroup.Consume(ctx, topics, handler); err != nil {
				log.Error("consumer error")
			}
		}
	}
}
