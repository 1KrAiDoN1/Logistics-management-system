package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

type KafkaConfig struct {
	Brokers  []string
	Topic    string
	Group_id string
}

type Msg struct {
	OrderId int64
	Driver  Driver
}

type Driver struct {
	ID            int64  `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Phone         string `json:"phone" db:"phone"`
	LicenseNumber string `json:"license_number" db:"license_number"`
	Car           string `json:"car" db:"car"`
}

func EnsureTopicExists(ctx context.Context, cfg KafkaConfig, logger *slog.Logger) error {
	// Используем первый брокер из списка
	if len(cfg.Brokers) == 0 {
		return fmt.Errorf("no brokers configured")
	}

	conn, err := kafka.DialContext(ctx, "tcp", cfg.Brokers[0])
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
		if p.Topic == cfg.Topic {
			topicExists = true
			break
		}
	}

	if topicExists {
		if logger != nil {
			logger.Info("Topic already exists", slog.String("topic", cfg.Topic))
		}
		return nil
	}

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
		Topic:             cfg.Topic,
		NumPartitions:     1, // Настройте по необходимости
		ReplicationFactor: 1, // Настройте по необходимости
	}

	err = controllerConn.CreateTopics(topicConfig)
	if err != nil {
		return fmt.Errorf("failed to create topic: %w", err)
	}

	return nil
}
