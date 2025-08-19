package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
	ctx    context.Context
}

type RedisConfig struct {
	Address      string `yaml:"address"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
	MaxRetries   int    `yaml:"max_retries"`
	DialTimeout  int    `yaml:"dial_timeout_seconds"`
	ReadTimeout  int    `yaml:"read_timeout_seconds"`
	WriteTimeout int    `yaml:"write_timeout_seconds"`
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
		DialTimeout:  time.Duration(cfg.DialTimeout),
		ReadTimeout:  time.Duration(cfg.ReadTimeout),
		WriteTimeout: time.Duration(cfg.WriteTimeout),
	})

	// Проверяем подключение
	_, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Println("Ошибка подключения к Redis:", err)
		return nil, err
	}
	fmt.Println("Подключение к Redis успешно")
	err = client.Set(ctx, "example_key", "example_value", 0).Err()
	if err != nil {
		fmt.Println("Ошибка при установке значения в Redis:", err)
		return nil, err
	}
	res, err := client.Get(ctx, "example_key").Result()
	if err != nil {
		fmt.Println("Ошибка при получении значения из Redis:", err)
		return nil, err
	}
	fmt.Println("Полученное значение из Redis:", res)
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
