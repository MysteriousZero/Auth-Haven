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

type AuthHandler struct {
	authService interfaces.AuthService
	validator   *validation.CustomValidator
}

func NewAuthHandler(authService interfaces.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validation.NewCustomValidator(),
	}
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": validation.GetValidationErrorMessage(err),
		})
		return
	}

	// Set client info in context
	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	result, err := h.authService.Login(ctx, req.TenantID, req.Email, req.Password)
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// VerifyMFA handles MFA verification
func (h *AuthHandler) VerifyMFA(c *gin.Context) {
	var req models.VerifyMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": validation.GetValidationErrorMessage(err),
		})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	result, err := h.authService.VerifyMFA(ctx, req.TempToken, req.Code)
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": validation.GetValidationErrorMessage(err),
		})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	tokens, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": "Session ID required",
		})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	err := h.authService.Logout(ctx, sessionID)
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// LogoutAll handles logout from all devices
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// Get user ID from JWT token (would be set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Authentication required",
		})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	err := h.authService.LogoutAll(ctx, userID.(string))
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// Registration handlers
type RegistrationHandler struct {
	registrationService interfaces.RegistrationService
	validator           *validation.CustomValidator
}

func NewRegistrationHandler(registrationService interfaces.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{
		registrationService: registrationService,
		validator:           validation.NewCustomValidator(),
	}
}

// RegisterIndividual handles individual user registration
func (h *RegistrationHandler) RegisterIndividual(c *gin.Context) {
	var req models.RegisterIndividualRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": validation.GetValidationErrorMessage(err),
		})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	user, err := h.registrationService.RegisterIndividual(ctx, req.Email, req.Password, req.FullName)
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// RegisterOrgUser handles organization user registration
func (h *RegistrationHandler) RegisterOrgUser(c *gin.Context) {
	var req models.RegisterOrgUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": validation.GetValidationErrorMessage(err),
		})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	user, err := h.registrationService.RegisterOrgUser(ctx, req.Email, req.Password, req.FullName)
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// RegisterWithInvitation handles registration via invitation
func (h *RegistrationHandler) RegisterWithInvitation(c *gin.Context) {
	var req models.RegisterWithInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	if err := h.validator.ValidateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ValidationError",
			"message": validation.GetValidationErrorMessage(err),
		})
		return
	}

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	user, err := h.registrationService.RegisterWithInvitation(ctx, req.InvitationToken, req.Password, req.FullName)
	if err != nil {
		status, errorCode := errors.HTTPError(err)
		c.JSON(status, gin.H{
			"error":   errorCode,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}
