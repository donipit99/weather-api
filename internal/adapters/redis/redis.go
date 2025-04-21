package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func NewClient(addr string) *Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // Без пароля
		DB:       0,  // База по умолчанию
	})
	return &Client{client: client}
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.client.Ping(ctx).Result()
}

func (c *Client) Set(ctx context.Context, key string, value interface{}) error {
	return c.client.Set(ctx, key, value, 5*time.Minute).Err() // TTL 5 минут
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *Client) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.client.TTL(ctx, key).Result()
}
