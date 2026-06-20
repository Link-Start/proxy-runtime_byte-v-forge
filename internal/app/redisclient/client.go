package redisclient

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, rawURL string) (*redis.Client, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("PROXY_RUNTIME_REDIS_URL is required")
	}
	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, errors.New("parse redis url: invalid redis url")
	}
	client := redis.NewClient(opts)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}

func Key(prefix string, value string) (string, bool) {
	prefix = strings.Trim(strings.TrimSpace(prefix), ":")
	value = strings.TrimSpace(value)
	if value == "" {
		return "", false
	}
	if prefix == "" {
		return value, true
	}
	return prefix + ":" + value, true
}
