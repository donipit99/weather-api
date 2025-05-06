package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
	ttl    time.Duration
}

func NewClient(addr string, ttlSeconds int) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // Без пароля
		DB:       0,  // База по умолчанию
	})
	return &Client{
		client: client,
		ttl:    time.Duration(ttlSeconds) * time.Second,
	}
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.Ping(ctx).Result()
}

func (c *Client) Set(ctx context.Context, key string, value any) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.Set(ctx, key, value, c.ttl).Err()
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.Del(ctx, key).Err()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.TTL(ctx, key).Result()
}
