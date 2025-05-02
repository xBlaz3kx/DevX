package tests

import (
	"context"

	"github.com/redis/go-redis/v9"
	redisT "github.com/testcontainers/testcontainers-go/modules/redis"
)

type RedisContainer struct {
	*redisT.RedisContainer
}

func NewRedisContainer(ctx context.Context) (*RedisContainer, error) {
	// Spin up a test redis container
	redisContainer, err := redisT.Run(ctx, "redis:7")
	if err != nil {
		return nil, err
	}

	return &RedisContainer{RedisContainer: redisContainer}, nil
}

func (r *RedisContainer) CreateRedisClient(ctx context.Context) (*redis.Client, error) {
	return CreateRedisClient(ctx, r.RedisContainer)
}

func CreateRedisClient(ctx context.Context, container *redisT.RedisContainer) (*redis.Client, error) {
	connectionString, err := container.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	url, err := redis.ParseURL(connectionString)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(url)
	return redisClient, nil
}
