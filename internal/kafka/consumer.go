package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	Conn   *kafka.Conn
	reader *kafka.Reader
	logger *slog.Logger
}

func NewKafkaConsumer(log *slog.Logger, cfg KafkaConfig) *KafkaConsumer {
	var conn *kafka.Conn
	var err error

	// Задаем параметры для повторных попыток
	maxRetries := 5                   // Максимальное количество попыток
	initialBackoff := 2 * time.Second // Начальная задержка
	maxBackoff := 30 * time.Second    // Максимальная задержка

	currentBackoff := initialBackoff

	for attempt := 1; attempt <= maxRetries; attempt++ {
		log.Info("Попытка подключения к Kafka", slog.Int("attempt", attempt))

		// Для проверки соединения достаточно подключиться к лидеру партиции 0 топика.
		// Это легкая операция, которая подтверждает, что брокер доступен и топик существует.
		// Используем контекст с таймаутом для каждой попытки.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		conn, err = kafka.DialLeader(ctx, "tcp", cfg.Brokers[0], cfg.Topic, 0)
		if err == nil {
			// Успех!
			log.Info("Успешно подключено к Kafka")
			conn.Close() // Закрываем соединение, так как оно нам больше не нужно.
			break
		}

		log.Error("Не удалось подключиться к Kafka", slog.String("error", err.Error()), slog.Int("attempt", attempt))

		if attempt < maxRetries {
			log.Info("Ожидание перед следующей попыткой", slog.String("duration", currentBackoff.String()))
			time.Sleep(currentBackoff)
			// Увеличиваем задержку для следующего раза (экспоненциальный рост)
			currentBackoff *= 2
			if currentBackoff > maxBackoff {
				currentBackoff = maxBackoff // Ограничиваем максимальную задержку
			}
		}
	}

	// Если все попытки провалились, возвращаем последнюю ошибку.
	if err != nil {
		log.Error("Все попытки подключения к Kafka исчерпаны", slog.String("error", err.Error()))
		return nil
	}

	// Если подключение успешно, создаем и возвращаем KafkaConsumer
	fmt.Println("Successfully connected to Kafka after retries")
	return &KafkaConsumer{
		Conn: conn,
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:          cfg.Brokers,
			Topic:            cfg.Topic,
			GroupID:          cfg.Group_id,
			MinBytes:         1,
			MaxBytes:         10e4,
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
