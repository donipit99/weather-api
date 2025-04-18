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

func (c *Client) Set(ctx context.Context, key string, value interface{}) error {

	return c.client.Set(ctx, key, value, time.Minute).Err() // ttl 1m
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
