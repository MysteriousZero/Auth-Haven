package server

import (
	"auth-haven/internal/config"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func TestCORSMiddlewareRejectsUnconfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(corsMiddleware(config.CORSConfig{AllowedOrigins: []string{"https://app.example.com"}, AllowCredentials: true}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestReadinessFailsWhenRequiredDependenciesAreUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := sql.Open("postgres", "host=127.0.0.1 port=1 connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: 0})
	defer redisClient.Close()

	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/ready", nil)
	readinessHandler(db, redisClient, true)(context)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestDatabaseMetricsExposeConfiguredPoolMaximum(t *testing.T) {
	db, err := sql.Open("postgres", "host=127.0.0.1 port=1")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(17)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequest(http.MethodGet, "/metrics", nil)
	databaseMetricsHandler(db)(context)
	if got := response.Body.String(); !strings.Contains(got, "auth_haven_db_max_open_connections 17") {
		t.Fatalf("metrics = %q", got)
	}
}

func TestCORSMiddlewareAllowsConfiguredOriginAndCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(corsMiddleware(config.CORSConfig{AllowedOrigins: []string{"https://app.example.com"}, AllowCredentials: true}))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "https://app.example.com")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("allow origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow credentials = %q", got)
	}
}
