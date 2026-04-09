package handlers

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/utils"
	"auth-haven/internal/validation"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MFAHandler struct {
	mfaService interfaces.MFAService
	validator  *validation.CustomValidator
}

func NewMFAHandler(mfaService interfaces.MFAService) *MFAHandler {
	return &MFAHandler{
		mfaService: mfaService,
		validator:  validation.NewCustomValidator(),
	}
}

// EnrollTOTP handles POST /v1/me/mfa/totp/enroll
func (h *MFAHandler) EnrollTOTP(c *gin.Context) {
	userID, _ := c.Get("user_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	enrollment, err := h.mfaService.EnrollTOTP(ctx, userID.(string))
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, enrollment)
}

// ActivateTOTP handles POST /v1/me/mfa/totp/activate
func (h *MFAHandler) ActivateTOTP(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		MFAID string `json:"mfa_id" validate:"required,uuid"`
		Code  string `json:"code" validate:"required,len=6"`
	}
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

	if err := h.mfaService.ActivateTOTP(ctx, userID.(string), req.MFAID, req.Code); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListMFAMethods handles GET /v1/me/mfa
func (h *MFAHandler) ListMFAMethods(c *gin.Context) {
	userID, _ := c.Get("user_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	methods, err := h.mfaService.ListMFAMethods(ctx, userID.(string))
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"methods": methods})
}

// DisableMFAMethod handles DELETE /v1/me/mfa/:mfa_id (soft disable)
func (h *MFAHandler) DisableMFAMethod(c *gin.Context) {
	userID, _ := c.Get("user_id")
	mfaID := c.Param("mfa_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.mfaService.DisableMFAMethod(ctx, userID.(string), mfaID); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteMFAMethod handles DELETE /v1/me/mfa/:mfa_id/delete (hard delete)
func (h *MFAHandler) DeleteMFAMethod(c *gin.Context) {
	userID, _ := c.Get("user_id")
	mfaID := c.Param("mfa_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.mfaService.DeleteMFAMethod(ctx, userID.(string), mfaID); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
