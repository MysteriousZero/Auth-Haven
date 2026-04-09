package service

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"auth-haven/internal/utils"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type sessionService struct {
	authRepo   interfaces.AuthRepository
	deviceRepo interfaces.DeviceRepository
	userRepo   interfaces.UserRepository
	auditRepo  interfaces.AuditRepository
}

func NewSessionService(
	authRepo interfaces.AuthRepository,
	deviceRepo interfaces.DeviceRepository,
	userRepo interfaces.UserRepository,
	auditRepo interfaces.AuditRepository,
) interfaces.SessionService {
	return &sessionService{
		authRepo:   authRepo,
		deviceRepo: deviceRepo,
		userRepo:   userRepo,
		auditRepo:  auditRepo,
	}
}

func (s *sessionService) ListSessions(ctx context.Context, userID string, limit, offset int) ([]models.Session, error) {
	sessions, err := s.authRepo.ListSessions(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	return sessions, nil
}

func (s *sessionService) RevokeSession(ctx context.Context, actorID, sessionID string) error {
	// Get session to verify ownership
	session, err := s.authRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Check if actor owns the session or is admin
	if session.UserID != actorID {
		// For now, we'll allow users to revoke their own sessions only
		// In a real implementation, you'd check admin permissions
		return errors.ErrForbidden
	}

	// Revoke the session
	err = s.authRepo.RevokeSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		Action:    models.AuditSessionRevoked,
		TargetID:  &sessionID,
		Metadata:  map[string]interface{}{"session_id": sessionID},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

func (s *sessionService) ListDevices(ctx context.Context, userID string, limit, offset int) ([]models.Device, error) {
	devices, err := s.deviceRepo.ListDevices(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return devices, nil
}

func (s *sessionService) RemoveDevice(ctx context.Context, actorID, deviceID string) error {
	// Get device to verify ownership
	device, err := s.deviceRepo.GetDeviceByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	// Check if actor owns the device
	if device.UserID != actorID {
		return errors.ErrForbidden
	}

	// Get user to get tenant ID for audit log
	user, err := s.userRepo.GetUserByID(ctx, device.UserID)
	if err != nil {
		// For audit log, we'll use a fallback
		user = &models.User{TenantID: "unknown"}
	}

	// Get all sessions for this device and revoke them
	sessions, err := s.authRepo.ListSessionsByDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device sessions: %w", err)
	}

	// Revoke all sessions for this device
	for _, session := range sessions {
		err = s.authRepo.RevokeSession(ctx, session.SessionID)
		if err != nil {
			fmt.Printf("Failed to revoke session %s: %v\n", session.SessionID, err)
		}
	}

	// Remove the device
	err = s.deviceRepo.DeleteDevice(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("failed to remove device: %w", err)
	}

	// Record audit log
	auditLog := &models.AuditLog{
		LogID:     uuid.New().String(),
		UserID:    &actorID,
		TenantID:  user.TenantID,
		Action:    models.AuditDeviceRemoved,
		TargetID:  &deviceID,
		Metadata:  map[string]interface{}{"device_id": deviceID, "device_name": device.DeviceName},
		IPAddress: utils.GetClientIP(ctx),
		UserAgent: utils.GetUserAgent(ctx),

	TraceID:   utils.GetTraceID(ctx),

	}

	err = s.auditRepo.CreateAuditLog(ctx, auditLog)
	if err != nil {
		fmt.Printf("Failed to create audit log: %v\n", err)
	}

	return nil
}

