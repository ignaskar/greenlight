package limiter

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	rdb           *redis.Client
	maxRequests   int64
	fixedWindow   time.Duration
	counterWindow time.Duration
}

func NewRateLimiter(rdb *redis.Client, maxRequests int64, fixedWindow time.Duration) *RateLimiter {
	if fixedWindow < 1*time.Hour {
		panic("rate limiter fixed window must be at least an hour")
	}

	counterWindow := fixedWindow / 60

	return &RateLimiter{
		rdb:           rdb,
		maxRequests:   maxRequests,
		fixedWindow:   fixedWindow,
		counterWindow: counterWindow,
	}
}

func (rl *RateLimiter) getCurrentWindowKey(identifier string) string {
	return fmt.Sprintf("rate_limit:%s", identifier)
}

func (rl *RateLimiter) getCounterField() string {
	windowSeconds := int64(rl.counterWindow.Seconds())
	return strconv.FormatInt(time.Now().Unix()/windowSeconds, 10)
}

func (rl *RateLimiter) cleanupOldWindows(ctx context.Context, key string) error {
	oldestValidTimestamp := time.Now().Add(-rl.fixedWindow).Unix() / int64(rl.counterWindow.Seconds())

	fields, err := rl.rdb.HKeys(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get hash keys: %w", err)
	}

	var expired []string
	for _, field := range fields {
		timestamp, _ := strconv.ParseInt(field, 10, 64)
		if timestamp < oldestValidTimestamp {
			expired = append(expired, field)
		}
	}

	if len(expired) > 0 {
		if err := rl.rdb.HDel(ctx, key, expired...).Err(); err != nil {
			return fmt.Errorf("failed to remove expired fields: %w", err)
		}
	}

	return nil
}

func (rl *RateLimiter) IsRateLimited(ctx context.Context, identifier string) (bool, error) {
	key := rl.getCurrentWindowKey(identifier)
	currentCounter := rl.getCounterField()

	pipe := rl.rdb.Pipeline()
	incr := pipe.HIncrBy(ctx, key, currentCounter, 1)
	pipe.Expire(ctx, key, rl.fixedWindow+rl.counterWindow)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("failed to increment counter: %w", err)
	}

	if incr.Val()%100 == 0 {
		if err := rl.cleanupOldWindows(ctx, key); err != nil {
			return false, err
		}
	}

	counters, err := rl.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to get counters: %w", err)
	}

	oldestValidTimestamp := time.Now().Add(-rl.fixedWindow).Unix() / int64(rl.counterWindow.Seconds())
	currentTimestamp := time.Now().Unix()
	if currentTimestamp%int64(rl.counterWindow.Seconds()) != 0 {
		oldestValidTimestamp--
	}

	var total int64
	for timestampStr, count := range counters {
		timestamp, _ := strconv.ParseInt(timestampStr, 10, 64)
		if timestamp >= oldestValidTimestamp {
			countVal, _ := strconv.ParseInt(count, 10, 64)
			total += countVal
		}
	}

	return total > rl.maxRequests, nil
}
