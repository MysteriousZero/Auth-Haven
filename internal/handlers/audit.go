package handlers

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AuditHandler struct {
	auditService interfaces.AuditService
}

func NewAuditHandler(auditService interfaces.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

// ListTenantLogs handles GET /v1/tenants/:tenant_id/audit-logs
func (h *AuditHandler) ListTenantLogs(c *gin.Context) {
	actorID, _ := c.Get("user_id")
	tenantID := c.Param("tenant_id")
	cursor := c.Query("cursor")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	logs, nextCursor, err := h.auditService.ListTenantLogs(ctx, actorID.(string), tenantID, cursor, limit)
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	resp := gin.H{
		"logs":  logs,
		"limit": limit,
	}
	if nextCursor != "" {
		resp["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, resp)
}

// ListUserLogs handles GET /v1/me/audit-logs
func (h *AuditHandler) ListUserLogs(c *gin.Context) {
	actorID, _ := c.Get("user_id")
	cursor := c.Query("cursor")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	userID := actorID.(string)
	logs, nextCursor, err := h.auditService.ListUserLogs(ctx, userID, userID, cursor, limit)
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	resp := gin.H{
		"logs":  logs,
		"limit": limit,
	}
	if nextCursor != "" {
		resp["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, resp)
}
