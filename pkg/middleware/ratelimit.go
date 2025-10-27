package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RateLimiter handles rate limiting using Redis
type RateLimiter struct {
	redis      *redis.Client
	maxRequests int
	window     time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisClient *redis.Client, maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		redis:      redisClient,
		maxRequests: maxRequests,
		window:     window,
	}
}

// Middleware returns a Fiber middleware for rate limiting
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client identifier (IP or user ID if authenticated)
		identifier := c.IP()

		// Check if user is authenticated and use user ID instead
		if userID := c.Locals("user_id"); userID != nil {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		key := fmt.Sprintf("ratelimit:%s", identifier)

		ctx := context.Background()

		// Increment counter
		count, err := rl.redis.Incr(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow the request but log the error
			return c.Next()
		}

		// Set expiration on first request
		if count == 1 {
			rl.redis.Expire(ctx, key, rl.window)
		}

		// Check if rate limit exceeded
		if count > int64(rl.maxRequests) {
			// Get TTL to inform client when they can retry
			ttl, _ := rl.redis.TTL(ctx, key).Result()

			c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(ttl).Unix()))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
				"retry_after": int(ttl.Seconds()),
			})
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxRequests))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.maxRequests-int(count)))

		return c.Next()
	}
}

// SimpleRateLimiter provides in-memory rate limiting (for development)
type SimpleRateLimiter struct {
	requests map[string][]time.Time
	max      int
	window   time.Duration
}

// NewSimpleRateLimiter creates a simple in-memory rate limiter
func NewSimpleRateLimiter(max int, window time.Duration) *SimpleRateLimiter {
	return &SimpleRateLimiter{
		requests: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}
}

// Middleware returns a Fiber middleware for simple rate limiting
func (srl *SimpleRateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		identifier := c.IP()

		if userID := c.Locals("user_id"); userID != nil {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		now := time.Now()

		// Clean old requests
		if requests, ok := srl.requests[identifier]; ok {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < srl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			srl.requests[identifier] = validRequests
		}

		// Check rate limit
		if len(srl.requests[identifier]) >= srl.max {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "rate limit exceeded",
			})
		}

		// Add current request
		srl.requests[identifier] = append(srl.requests[identifier], now)

		return c.Next()
	}
}
