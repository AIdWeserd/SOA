package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"flight-service/internal/repository"
)

var ErrCacheMiss = errors.New("cache miss")

type FlightCache struct {
	client    redis.UniversalClient
	ttlFlight time.Duration
	ttlSearch time.Duration
	logger    *zap.Logger
}

func New(client redis.UniversalClient, ttlFlightSec, ttlSearchSec int, logger *zap.Logger) *FlightCache {
	return &FlightCache{
		client:    client,
		ttlFlight: time.Duration(ttlFlightSec) * time.Second,
		ttlSearch: time.Duration(ttlSearchSec) * time.Second,
		logger:    logger,
	}
}

func flightKey(id string) string {
	return fmt.Sprintf("flight:%s", id)
}

func searchKey(origin, destination, date string) string {
	return fmt.Sprintf("search:%s:%s:%s", origin, destination, date)
}

func (c *FlightCache) GetFlight(ctx context.Context, id string) (*repository.FlightModel, error) {
	key := flightKey(id)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.logger.Debug("cache miss", zap.String("key", key))
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var f repository.FlightModel
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	c.logger.Debug("cache hit", zap.String("key", key))
	return &f, nil
}

func (c *FlightCache) SetFlight(ctx context.Context, f *repository.FlightModel) error {
	data, err := json.Marshal(f)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.client.Set(ctx, flightKey(f.ID), data, c.ttlFlight).Err()
}

func (c *FlightCache) DeleteFlight(ctx context.Context, id string) error {
	return c.client.Del(ctx, flightKey(id)).Err()
}

func (c *FlightCache) GetSearch(ctx context.Context, origin, destination, date string) ([]*repository.FlightModel, error) {
	key := searchKey(origin, destination, date)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			c.logger.Debug("cache miss", zap.String("key", key))
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("redis get: %w", err)
	}

	var flights []*repository.FlightModel
	if err := json.Unmarshal(data, &flights); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	c.logger.Debug("cache hit", zap.String("key", key))
	return flights, nil
}

func (c *FlightCache) SetSearch(ctx context.Context, origin, destination, date string, flights []*repository.FlightModel) error {
	data, err := json.Marshal(flights)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	return c.client.Set(ctx, searchKey(origin, destination, date), data, c.ttlSearch).Err()
}

// DeleteSearchByFlight удаляет все ключи поиска для маршрута через SCAN,
// т.к. неизвестно, какие даты закешированы
func (c *FlightCache) DeleteSearchByFlight(ctx context.Context, flightOrigin, flightDest string) error {
	pattern := fmt.Sprintf("search:%s:%s:*", flightOrigin, flightDest)
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan: %w", err)
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("del: %w", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
