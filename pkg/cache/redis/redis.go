package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
	ctx    context.Context
}

type RedisConfig struct {
	Address      string `mapstructure:"address"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	MaxRetries   int    `mapstructure:"max_retries"`
	DialTimeout  int    `mapstructure:"dial_timeout_seconds"`
	ReadTimeout  int    `mapstructure:"read_timeout_seconds"`
	WriteTimeout int    `mapstructure:"write_timeout_seconds"`
}

func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
	ctx := context.Background()
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,      // Адрес сервера Redis
		Password:     cfg.Password,     // Пароль (если установлен)
		DB:           cfg.DB,           // Используемая база данных
		PoolSize:     cfg.PoolSize,     // 2x количество CPU cores
		MinIdleConns: cfg.MinIdleConns, // Половина от PoolSize
		MaxRetries:   cfg.MaxRetries,   // Количество повторов
		DialTimeout:  time.Duration(cfg.DialTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	})
	slog.Info("Redis connection attempt",
		slog.String("address", cfg.Address),
		slog.Int("dial_timeout", cfg.DialTimeout))

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := client.Ping(pingCtx).Result()
	if err != nil {
		slog.Error("Redis ping failed",
			slog.String("error", err.Error()),
			slog.String("address", cfg.Address))
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}

	slog.Info("Redis ping successful")
	slog.Info("Connected to Redis", slog.String("address", cfg.Address))

	redisClient := &RedisClient{
		Client: client,
		ctx:    ctx,
	}
	return redisClient, nil
}

// Close закрывает соединение
func (r *RedisClient) Close() error {
	return r.Client.Close()
}

// Ping проверяет соединение
func (r *RedisClient) Ping() (string, error) {
	return r.Client.Ping(r.ctx).Result()
}

// HealthCheck проверяет состояние Redis
func (r *RedisClient) HealthCheck() (*redis.PoolStats, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 2*time.Second)
	defer cancel()

	_, err := r.Client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis health check failed: %w", err)
	}

	// Проверяем статистику соединений
	stats := r.Client.PoolStats()

	return stats, nil
}
