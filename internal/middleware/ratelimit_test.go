package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func unavailableRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: time.Millisecond, ReadTimeout: time.Millisecond, WriteTimeout: time.Millisecond, MaxRetries: 0})
}

func TestRateLimiterFailsClosedWhenRedisIsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := unavailableRedis()
	defer client.Close()
	router := gin.New()
	router.POST("/login", NewRateLimiter(client, false).LoginRateLimit(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/login", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestRateLimiterCanFailOpenWhenExplicitlyConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	client := unavailableRedis()
	defer client.Close()
	router := gin.New()
	router.POST("/login", NewRateLimiter(client, true).LoginRateLimit(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/login", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("X-RateLimit-Status"); got != "degraded" {
		t.Fatalf("degraded header = %q", got)
	}
}
