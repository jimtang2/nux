package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	Client *redis.Client
}

func (c *Redis) Connect(connStr string) error {
	if connStr == "" {
		return fmt.Errorf("redis connection string is empty")
	}

	opt, err := redis.ParseURL(connStr)
	if err != nil {
		return err
	}

	client := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return err
	}

	c.Client = client
	return nil
}
