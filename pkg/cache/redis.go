package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps the Redis client with helper methods.
// If Redis is not configured or unreachable, all operations are no-ops
// so the app degrades gracefully without caching rather than crashing.
type Client struct {
	rdb     *redis.Client
	enabled bool
}

var ctx = context.Background()

// NewClient creates a Redis client from environment variables.
// Returns a disabled client if Redis is not configured.
func NewClient() *Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	password := os.Getenv("REDIS_PASSWORD")
	dbStr := os.Getenv("REDIS_DB")

	if host == "" || port == "" {
		return &Client{enabled: false}
	}

	db := 0
	if dbStr != "" {
		if n, err := strconv.Atoi(dbStr); err == nil {
			db = n
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     50,
		MinIdleConns: 5,
	})

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return &Client{enabled: false}
	}

	return &Client{rdb: rdb, enabled: true}
}

// IsEnabled returns whether Redis is available.
func (c *Client) IsEnabled() bool {
	return c.enabled
}

// Set stores a value as JSON with a TTL.
func (c *Client) Set(key string, value interface{}, ttl time.Duration) error {
	if !c.enabled {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, key, data, ttl).Err()
}

// Get retrieves a JSON value and unmarshals it into dest.
// Returns (false, nil) on cache miss, (true, nil) on hit.
func (c *Client) Get(key string, dest interface{}) (bool, error) {
	if !c.enabled {
		return false, nil
	}
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return false, nil // cache miss
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(data, dest)
}

// Delete removes one or more keys.
func (c *Client) Delete(keys ...string) error {
	if !c.enabled {
		return nil
	}
	return c.rdb.Del(ctx, keys...).Err()
}

// DeletePattern removes all keys matching a glob pattern.
func (c *Client) DeletePattern(pattern string) error {
	if !c.enabled {
		return nil
	}
	keys, err := c.rdb.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return err
	}
	return c.rdb.Del(ctx, keys...).Err()
}
