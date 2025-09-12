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
	config KafkaConfig
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
		config: cfg,
	}
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, msg kafka.Message) error {
	slog.Info("Attempting to send message to Kafka",
		slog.String("topic", kp.writer.Topic),
		slog.String("key", string(msg.Key)),
		slog.Int("value_size", len(msg.Value)))
	err := kp.writer.WriteMessages(ctx, msg)
	if err != nil {
		slog.Error("Failed to send message to Kafka", slog.String("error", err.Error()))
		return fmt.Errorf("failed to send message: %w", err)
	}
	slog.Info("Message sent to Kafka", slog.String("topic", kp.writer.Topic), slog.String("key", string(msg.Value)))

	return nil
}

func (kp *KafkaProducer) Close() error {
	return kp.writer.Close()
}

func (kp *KafkaProducer) HealthCheck(ctx context.Context) error {
	// Создаем временное соединение для проверки
	conn, err := kafka.DialContext(ctx, "tcp", kp.config.Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial Kafka: %w", err)
	}
	defer conn.Close()

	// Проверяем, что можем получить метаданные
	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to get brokers: %w", err)
	}

	return nil
}

func (kp *KafkaProducer) IsHealthy() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := kp.HealthCheck(ctx)
	return err == nil
}

// func CreateTopicIfNotExists() error {
// 	conn, err := kafka.Dial("tcp", "localhost:9092")
// 	if err != nil {
// 		return fmt.Errorf("не удалось подключиться к Kafka: %w", err)
// 	}
// 	defer conn.Close()

// 	// Получаем список топиков
// 	partitions, err := conn.ReadPartitions()
// 	if err != nil {
// 		return fmt.Errorf("не удалось получить список разделов: %w", err)
// 	}

// 	// Проверяем, существует ли топик
// 	topicExists := false
// 	for _, p := range partitions {
// 		if p.Topic == "drivers" {
// 			topicExists = true
// 			break
// 		}
// 	}

// 	if !topicExists {
// 		// Создаем топик
// 		controller, err := conn.Controller()
// 		if err != nil {
// 			return fmt.Errorf("не удалось получить контроллер: %w", err)
// 		}

// 		controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
// 		if err != nil {
// 			return fmt.Errorf("не удалось подключиться к контроллеру: %w", err)
// 		}
// 		defer controllerConn.Close()

// 		topicConfig := kafka.TopicConfig{
// 			Topic:             "drivers",
// 			NumPartitions:     1,
// 			ReplicationFactor: 1,
// 		}

// 		err = controllerConn.CreateTopics(topicConfig)
// 		if err != nil {
// 			return fmt.Errorf("не удалось создать топик: %w", err)
// 		}

// 		slog.Info("✅ Топик 'drivers' создан")
// 	} else {
// 		slog.Info("✅ Топик 'drivers' уже существует")
// 	}

// 	return nil
// }
