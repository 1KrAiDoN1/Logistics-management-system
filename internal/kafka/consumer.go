package kafka

import (
	"context"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	reader *kafka.Reader
	logger *slog.Logger
}

func NewKafkaConsumer(log *slog.Logger, cfg KafkaConfig) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:          cfg.Brokers,
			Topic:            cfg.Topic,
			GroupID:          "order-service-group",
			MinBytes:         1,
			MaxBytes:         10e6, // 10MB
			CommitInterval:   time.Second,
			StartOffset:      kafka.LastOffset,
			ReadBatchTimeout: 100 * time.Millisecond,
		}),
		logger: log,
	}
}

func (kc *KafkaConsumer) ReadMessage(ctx context.Context) (*kafka.Message, error) {
	kafkaMessage, err := kc.reader.ReadMessage(ctx)
	if err != nil {
		kc.logger.Error("Failed to read message", slog.String("status", "error"), slog.String("error", err.Error()))
		return nil, err
	}
	return &kafkaMessage, nil
}

func (kc *KafkaConsumer) CommitMessage(ctx context.Context, msg *kafka.Message) error {
	return kc.reader.CommitMessages(ctx, *msg)
}

func (kc *KafkaConsumer) Close() error {
	return kc.reader.Close()
}
