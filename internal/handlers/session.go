package handlers

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	sessionService interfaces.SessionService
}

func NewSessionHandler(sessionService interfaces.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

// ListSessions handles GET /v1/me/sessions
func (h *SessionHandler) ListSessions(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	sessions, err := h.sessionService.ListSessions(ctx, userID.(string), limit, offset)
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

// RevokeSession handles DELETE /v1/me/sessions/:session_id
func (h *SessionHandler) RevokeSession(c *gin.Context) {
	userID, _ := c.Get("user_id")
	sessionID := c.Param("session_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.sessionService.RevokeSession(ctx, userID.(string), sessionID); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListDevices handles GET /v1/me/devices
func (h *SessionHandler) ListDevices(c *gin.Context) {
	userID, _ := c.Get("user_id")

	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	devices, err := h.sessionService.ListDevices(ctx, userID.(string), limit, offset)
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"devices": devices})
}

// RemoveDevice handles DELETE /v1/me/devices/:device_id
func (h *SessionHandler) RemoveDevice(c *gin.Context) {
	userID, _ := c.Get("user_id")
	deviceID := c.Param("device_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.sessionService.RemoveDevice(ctx, userID.(string), deviceID); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
