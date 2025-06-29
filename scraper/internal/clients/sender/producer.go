package sender

import (
	"context"
	"fmt"
	"scraper/utils"
	"time"

	"github.com/Shopify/sarama"
	"scraper/internal/model/scraper"
)

type Producer struct {
	asyncProducer sarama.AsyncProducer
	codec         utils.JSONCodec
	topic         string
	dlqTopic      string
}

func NewProducer(brokers []string, topic, dlqTopic string, timeout time.Duration) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Return.Errors = true
	cfg.Producer.Timeout = timeout

	prod, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot create Kafka async producer: %w", err)
	}

	p := &Producer{asyncProducer: prod, codec: utils.JSONCodec{}, topic: topic, dlqTopic: dlqTopic}

	go func() {
		for errMsg := range prod.Errors() {
			fmt.Printf("producer failed to send message: %v\n", err)

			failedMsg := &sarama.ProducerMessage{
				Topic: dlqTopic,
				Key:   errMsg.Msg.Key,
				Value: errMsg.Msg.Value,
			}

			p.asyncProducer.Input() <- failedMsg
		}
	}()

	return p, nil
}

func (p *Producer) Updates(ctx context.Context, req *scraper.LinkUpdate, isFailed bool) error {
	data, err := p.codec.Marshal(req)
	if err != nil || isFailed {
		dlqData := fmt.Sprintf(`{"error:"marshal failed, "raw":"%v"}`, req)
		dlqMsg := &sarama.ProducerMessage{
			Topic: p.dlqTopic,
			Value: sarama.ByteEncoder(dlqData),
		}

		select {
		case p.asyncProducer.Input() <- dlqMsg:
			return fmt.Errorf("json marshal LinkUpdate failed; sent to DLQ: %w", err)
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(data),
	}

	select {
	case p.asyncProducer.Input() <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Producer) Close() error {
	return p.asyncProducer.Close()
}
