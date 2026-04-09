package handlers

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/utils"
	"auth-haven/internal/validation"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type InvitationHandler struct {
	invitationService interfaces.InvitationService
	validator         *validation.CustomValidator
}

func NewInvitationHandler(invitationService interfaces.InvitationService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		validator:         validation.NewCustomValidator(),
	}
}

// SendInvitation handles POST /v1/tenants/:tenant_id/invitations
func (h *InvitationHandler) SendInvitation(c *gin.Context) {
	actorID, _ := c.Get("user_id")
	tenantID := c.Param("tenant_id")

	var req struct {
		Email  string `json:"email" validate:"required,email"`
		RoleID int64  `json:"role_id" validate:"required"`
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

	invitation, err := h.invitationService.SendInvitation(ctx, actorID.(string), tenantID, req.Email, req.RoleID)
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, invitation)
}

// ListInvitations handles GET /v1/tenants/:tenant_id/invitations
func (h *InvitationHandler) ListInvitations(c *gin.Context) {
	tenantID := c.Param("tenant_id")

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

	invitations, err := h.invitationService.ListInvitations(ctx, tenantID, limit, offset)
	if err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"invitations": invitations})
}

// RevokeInvitation handles DELETE /v1/tenants/:tenant_id/invitations/:invitation_id
func (h *InvitationHandler) RevokeInvitation(c *gin.Context) {
	actorID, _ := c.Get("user_id")
	invitationID := c.Param("invitation_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.invitationService.RevokeInvitation(ctx, actorID.(string), invitationID); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ResendInvitation handles POST /v1/tenants/:tenant_id/invitations/:invitation_id/resend
func (h *InvitationHandler) ResendInvitation(c *gin.Context) {
	actorID, _ := c.Get("user_id")
	invitationID := c.Param("invitation_id")

	ctx := utils.SetClientIP(c.Request.Context(), utils.GetRealIP(c.Request))
	ctx = utils.SetUserAgent(ctx, c.GetHeader("User-Agent"))

	if err := h.invitationService.ResendInvitation(ctx, actorID.(string), invitationID); err != nil {
		status, code := errors.HTTPError(err)
		c.JSON(status, gin.H{"error": code, "message": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
