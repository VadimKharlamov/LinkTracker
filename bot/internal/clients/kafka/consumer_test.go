package kafka_test

import (
	"bot/internal/clients/kafka"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"bot/internal/model/bot"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestConsumer_Integration(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	req := testcontainers.ContainerRequest{
		Image:        "bitnami/kafka:latest",
		ExposedPorts: []string{"9092/tcp", "9093/tcp"},
		Env: map[string]string{
			"ALLOW_PLAINTEXT_LISTENER":                 "yes",
			"KAFKA_CFG_PROCESS_ROLES":                  "broker,controller",
			"KAFKA_CFG_NODE_ID":                        "1",
			"KAFKA_CFG_LISTENERS":                      "PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093",
			"KAFKA_CFG_ADVERTISED_LISTENERS":           "PLAINTEXT://localhost:9092",
			"KAFKA_CFG_LISTENER_SECURITY_PROTOCOL_MAP": "PLAINTEXT:PLAINTEXT,CONTROLLER:PLAINTEXT",
			"KAFKA_CFG_CONTROLLER_LISTENER_NAMES":      "CONTROLLER",
			"KAFKA_CFG_CONTROLLER_QUORUM_VOTERS":       "1@localhost:9093",
			"KAFKA_INTER_BROKER_LISTENER_NAME":         "PLAINTEXT",
		},
		WaitingFor: wait.ForListeningPort("9092/tcp"),
	}

	kafkaC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := kafkaC.Host(ctx)

	require.NoError(t, err)

	mappedPort, err := kafkaC.MappedPort(ctx, "9092")

	require.NoError(t, err)

	brokerAddr := fmt.Sprintf("%s:%s", host, mappedPort.Port())
	brokers := []string{brokerAddr}

	prodCfg := sarama.NewConfig()

	prodCfg.Producer.Return.Successes = true

	prod, err := sarama.NewSyncProducer(brokers, prodCfg)

	require.NoError(t, err)

	defer prod.Close()

	topic := "link-updates-test"

	gotCh := make(chan *bot.LinkUpdate, 1)

	cctx, cancel := context.WithCancel(ctx)
	go func() {
		_ = kafka.RunConsumerGroup(
			cctx,
			logger,
			brokers,
			"test-group",
			[]string{topic},
			func(upd *bot.LinkUpdate) error {
				gotCh <- upd
				return nil
			},
		)
	}()

	time.Sleep(2 * time.Second)

	want := &bot.LinkUpdate{
		ID:          7,
		URL:         "https://go.dev",
		Description: "golang",
		TgChatIDs:   []int64{42, 43},
	}
	data, _ := json.Marshal(want)

	_, _, err = prod.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	})

	require.NoError(t, err)

	select {
	case got := <-gotCh:
		require.Equal(t, want, got)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for consumer to receive message")
	}

	cancel()
}
