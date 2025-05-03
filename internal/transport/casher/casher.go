package casher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const LIFE_TIME = 10 * time.Hour

type Casher struct {
	redisClient *redis.Client
}

func Init(redisClient *redis.Client) *Casher {
	return &Casher{
		redisClient: redisClient,
	}
}

func (c *Casher) DoCashing(ctx context.Context, key string, payload any) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res := c.redisClient.Set(ctx, key, jsonPayload, LIFE_TIME)
	if err = res.Err(); err != nil {
		return err
	}

	return nil
}
