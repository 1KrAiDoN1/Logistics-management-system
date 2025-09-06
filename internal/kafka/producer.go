package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"logistics/internal/shared/entity"
	"time"

	"github.com/segmentio/kafka-go"
)

type DriverFoundMessage struct {
	OrderID   int64          `json:"order_id"`
	Driver    *entity.Driver `json:"driver,omitempty"`
	Success   bool           `json:"success"`
	Message   string         `json:"message"`
	Timestamp int64          `json:"timestamp"`
}

// KafkaProducer структура для продюсера
type KafkaProducer struct {
	writer *kafka.Writer
}

func NewKafkaProducer(cfg KafkaConfig) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(cfg.Brokers...),
			Topic:        cfg.Topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireOne,
			Compression:  kafka.Snappy,
			BatchTimeout: 10 * time.Millisecond, // Быстрее отправка батчей
			WriteTimeout: 5 * time.Second,
		},
	}
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, msg kafka.Message) error {

	err := kp.writer.WriteMessages(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}

func CreateTopicIfNotExists() error {
	conn, err := kafka.Dial("tcp", "localhost:9092")
	if err != nil {
		return fmt.Errorf("не удалось подключиться к Kafka: %w", err)
	}
	defer conn.Close()

	// Получаем список топиков
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("не удалось получить список разделов: %w", err)
	}

	// Проверяем, существует ли топик
	topicExists := false
	for _, p := range partitions {
		if p.Topic == "drivers" {
			topicExists = true
			break
		}
	}

	if !topicExists {
		// Создаем топик
		controller, err := conn.Controller()
		if err != nil {
			return fmt.Errorf("не удалось получить контроллер: %w", err)
		}

		controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
		if err != nil {
			return fmt.Errorf("не удалось подключиться к контроллеру: %w", err)
		}
		defer controllerConn.Close()

		topicConfig := kafka.TopicConfig{
			Topic:             "drivers",
			NumPartitions:     1,
			ReplicationFactor: 1,
		}

		err = controllerConn.CreateTopics(topicConfig)
		if err != nil {
			return fmt.Errorf("не удалось создать топик: %w", err)
		}

		slog.Info("✅ Топик 'drivers' создан")
	} else {
		slog.Info("✅ Топик 'drivers' уже существует")
	}

	return nil
}
