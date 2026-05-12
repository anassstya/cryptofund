package prices

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func PopularPairs() []string {
	return []string{
		"BTC",
		"ETH",
		"BNB",
		"SOL",
		"XRP",
		"DOGE",
		"ADA",
		"TRX",
		"TON",
		"LINK",
	}
}

type PriceCache struct {
	client   *redis.Client
	ttl      time.Duration
	interval time.Duration
}

func NewPriceCache(client *redis.Client) *PriceCache {
	return &PriceCache{
		client:   client,
		ttl:      time.Minute * 4,
		interval: time.Minute * 3,
	}
}

func PriceCacheKey(exchangeName, pair string) string {
	return fmt.Sprintf("price:%s:%s", strings.ToLower(exchangeName), strings.ToUpper(pair))
}

func (p *PriceCache) SetPrice(ctx context.Context, key string, value float64) error {
	err := p.client.Set(ctx, key, value, p.ttl).Err()
	if err != nil {
		return fmt.Errorf("error setting in redis: %w", err)
	}
	return nil
}

func (p *PriceCache) GetPrice(ctx context.Context, key string) (float64, error) {
	value, err := p.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, redis.Nil
		}

		return 0, fmt.Errorf("error getting price from redis: %w", err)
	}

	price, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing price from redis: %w", err)
	}

	return price, nil
}
