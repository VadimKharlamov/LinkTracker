package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"bot/internal/model/bot"
	"github.com/gomodule/redigo/redis"
)

type Storage struct {
	pool *redis.Pool
}

func New(address string, maxIdle, maxActive int) *Storage {
	return &Storage{
		pool: &redis.Pool{
			MaxIdle:     maxIdle,
			MaxActive:   maxActive,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", address)
			},
			TestOnBorrow: func(c redis.Conn, _ time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

func (r *Storage) Close() error {
	return r.pool.Close()
}

func (r *Storage) SetLinks(ctx context.Context, key string, links []bot.Link) error {
	const op = "storage.redis.SetLinks"

	data, err := json.Marshal(links)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer conn.Close()

	if _, err = conn.Do("SET", key, data); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *Storage) GetLinks(ctx context.Context, key string) ([]bot.Link, error) {
	const op = "storage.redis.GetLinks"

	conn, err := r.pool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("GetLinks: pool.Get: %w", err)
	}
	defer conn.Close()

	raw, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if errors.Is(err, redis.ErrNil) {
			return nil, redis.ErrNil
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var links []bot.Link

	if err = json.Unmarshal(raw, &links); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return links, nil
}
