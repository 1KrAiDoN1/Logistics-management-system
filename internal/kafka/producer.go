package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"logistics/internal/shared/entity"
	"net"
	"strconv"
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
	Conn   *kafka.Conn
	writer *kafka.Writer
	config KafkaConfig
	logger *slog.Logger
}

func NewKafkaProducer(cfg KafkaConfig, log *slog.Logger) *KafkaProducer {
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

	return &KafkaProducer{
		Conn: conn,
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
		logger: log,
	}
}

func (kp *KafkaProducer) SendMessage(ctx context.Context, msg kafka.Message) error {
	kp.logger.Info("Attempting to send message to Kafka",
		slog.String("topic", kp.writer.Topic),
		slog.String("key", string(msg.Key)),
		slog.Int("value_size", len(msg.Value)))
	err := kp.writer.WriteMessages(ctx, msg)
	if err != nil {
		kp.logger.Error("Failed to send message to Kafka", slog.String("error", err.Error()))
		return fmt.Errorf("failed to send message: %w", err)
	}
	kp.logger.Info("Message sent to Kafka", slog.String("topic", kp.writer.Topic), slog.String("key", string(msg.Value)))

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

func (kp *KafkaProducer) EnsureTopicExists(ctx context.Context) error {
	conn, err := kafka.DialContext(ctx, "tcp", kp.config.Brokers[0])
	if err != nil {
		return fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	// Получаем метаданные для проверки существования топика
	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	// Проверяем существование топика
	topicExists := false
	for _, p := range partitions {
		if p.Topic == kp.config.Topic {
			topicExists = true
			break
		}
	}

	if !topicExists {
		// Получаем контроллер для создания топика
		controller, err := conn.Controller()
		if err != nil {
			return fmt.Errorf("failed to get controller: %w", err)
		}

		controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
		if err != nil {
			return fmt.Errorf("failed to connect to controller: %w", err)
		}
		defer controllerConn.Close()

		topicConfig := kafka.TopicConfig{
			Topic:             kp.config.Topic,
			NumPartitions:     1, // Настройте по необходимости
			ReplicationFactor: 1, // Настройте по необходимости
		}

		err = controllerConn.CreateTopics(topicConfig)
		if err != nil {
			return fmt.Errorf("failed to create topic: %w", err)
		}

		kp.logger.Info("Topic created", slog.String("topic", kp.config.Topic))
	}

	return nil
}
