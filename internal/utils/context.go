package utils

import (
	"context"
	"net"
	"net/http"
)

func GetClientIP(ctx context.Context) string {
	// Try to get IP from context first
	if ip, ok := ctx.Value("client_ip").(string); ok {
		return ip
	}

	// Fallback to localhost for now
	return "127.0.0.1"
}

func GetUserAgent(ctx context.Context) string {
	// Try to get user agent from context first
	if ua, ok := ctx.Value("user_agent").(string); ok {
		return ua
	}

	// Fallback
	return "auth-haven-internal"
}

func SetClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, "client_ip", ip)
}

func SetUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, "user_agent", userAgent)
}

func GetTraceID(ctx context.Context) string {
	if tid, ok := ctx.Value("trace_id").(string); ok {
		return tid
	}
	return ""
}

func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, "trace_id", traceID)
}

// Helper to extract real IP from request
func GetRealIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		for i, char := range xff {
			if char == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}
