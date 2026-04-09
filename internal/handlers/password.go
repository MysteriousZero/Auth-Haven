package handlers

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/utils"
	"auth-haven/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PasswordHandler struct {
	passwordService interfaces.PasswordService
	validator       *validation.CustomValidator
}

func NewPasswordHandler(passwordService interfaces.PasswordService) *PasswordHandler {
	return &PasswordHandler{
		passwordService: passwordService,
		validator:       validation.NewCustomValidator(),
	}
}

// ForgotPassword handles POST /v1/auth/password/forgot
func (h *PasswordHandler) ForgotPassword(c *gin.Context) {
	var req models.RequestPasswordResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidationError", "message": err.Error()})
		return
	}
	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidationError", "message": validation.GetValidationErrorMessage(err)})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	// Always 204 regardless of whether the email exists — prevents enumeration
	_ = h.passwordService.RequestPasswordReset(ctx, req.TenantID, req.Email)
	c.Status(http.StatusNoContent)
}

// ResetPassword handles POST /v1/auth/password/reset
func (h *PasswordHandler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidationError", "message": err.Error()})
		return
	}
	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidationError", "message": validation.GetValidationErrorMessage(err)})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.passwordService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ChangePassword handles POST /v1/auth/password/change (requires auth)
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidationError", "message": err.Error()})
		return
	}
	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ValidationError", "message": validation.GetValidationErrorMessage(err)})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.passwordService.ChangePassword(ctx, userID.(string), req.CurrentPassword, req.NewPassword); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
