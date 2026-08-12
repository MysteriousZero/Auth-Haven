package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	redisClient *redis.Client
	failOpen    bool
}

func NewRateLimiter(redisClient *redis.Client, failOpen ...bool) *RateLimiter {
	allowOnFailure := false
	if len(failOpen) > 0 {
		allowOnFailure = failOpen[0]
	}
	return &RateLimiter{
		redisClient: redisClient,
		failOpen:    allowOnFailure,
	}
}

func (rl *RateLimiter) handleFailure(c *gin.Context, err error) bool {
	if rl.failOpen {
		c.Header("X-RateLimit-Status", "degraded")
		return true
	}
	c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable"})
	c.Abort()
	return false
}

// Rate limiting configurations
var (
	// Login: 5 attempts per 5 minutes per IP, 10 attempts per 5 minutes per email
	loginLimits = map[string]RateLimit{
		"ip":    {Attempts: 5, Window: 5 * time.Minute},
		"email": {Attempts: 10, Window: 5 * time.Minute},
	}

	// Password forgot: 3 attempts per 15 minutes per IP
	passwordForgotLimits = map[string]RateLimit{
		"ip": {Attempts: 3, Window: 15 * time.Minute},
	}

	// MFA verify: 5 attempts per 15 minutes per IP
	mfaLimits = map[string]RateLimit{
		"ip": {Attempts: 5, Window: 15 * time.Minute},
	}
)

type RateLimit struct {
	Attempts int
	Window   time.Duration
}

func (rl *RateLimiter) CheckRateLimit(ctx context.Context, key string, limit RateLimit) (bool, time.Duration, error) {
	current, err := rl.redisClient.Get(ctx, key).Int()
	if err != nil && err != redis.Nil {
		return false, 0, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if current >= limit.Attempts {
		// Get TTL for remaining time
		ttl, err := rl.redisClient.TTL(ctx, key).Result()
		if err != nil {
			return false, 0, fmt.Errorf("failed to get TTL: %w", err)
		}
		return false, ttl, nil
	}

	// Increment counter
	pipe := rl.redisClient.Pipeline()
	pipe.Incr(ctx, key)
	if current == 0 {
		// Set expiration only on first increment
		pipe.Expire(ctx, key, limit.Window)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return true, 0, nil
}

func (rl *RateLimiter) LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		email := ""

		// Attempt to get email from JSON body (LoginRequest)
		if c.Request.Method == http.MethodPost {
			// Read body
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore body for the main handler
				c.Request.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

				var loginReq struct {
					Email string `json:"email"`
				}
				if json.Unmarshal(bodyBytes, &loginReq) == nil {
					email = loginReq.Email
				}
			}
		}

		// Fallback to PostForm if email is still empty
		if email == "" {
			email = c.PostForm("email")
		}

		// Check IP rate limit
		allowed, retryAfter, err := rl.CheckRateLimit(c.Request.Context(),
			fmt.Sprintf("rate_limit:login:ip:%s", ip), loginLimits["ip"])
		if err != nil {
			if rl.handleFailure(c, err) {
				c.Next()
			}
			return
		}

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "RateLimitExceeded",
				"message": "Too many login attempts. Please try again later.",
			})
			c.Abort()
			return
		}

		// Check email rate limit if email is provided
		if email != "" {
			allowed, retryAfter, err = rl.CheckRateLimit(c.Request.Context(),
				fmt.Sprintf("rate_limit:login:email:%s", strings.ToLower(email)), loginLimits["email"])
			if err != nil {
				if rl.handleFailure(c, err) {
					c.Next()
				}
				return
			}

			if !allowed {
				c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "RateLimitExceeded",
					"message": "Too many login attempts for this email. Please try again later.",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func (rl *RateLimiter) PasswordForgotRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		allowed, retryAfter, err := rl.CheckRateLimit(c.Request.Context(),
			fmt.Sprintf("rate_limit:password_forgot:ip:%s", ip), passwordForgotLimits["ip"])
		if err != nil {
			if rl.handleFailure(c, err) {
				c.Next()
			}
			return
		}

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "RateLimitExceeded",
				"message": "Too many password reset requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (rl *RateLimiter) MFAVerifyRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		allowed, retryAfter, err := rl.CheckRateLimit(c.Request.Context(),
			fmt.Sprintf("rate_limit:mfa:ip:%s", ip), mfaLimits["ip"])
		if err != nil {
			if rl.handleFailure(c, err) {
				c.Next()
			}
			return
		}

		if !allowed {
			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "RateLimitExceeded",
				"message": "Too many MFA verification attempts. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
