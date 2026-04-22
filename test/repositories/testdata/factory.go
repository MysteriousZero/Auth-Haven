package testdata

import (
	"time"

	"auth-haven/internal/domain/models"

	"github.com/google/uuid"
)

// CreateTestTenant creates a test tenant with realistic data
func CreateTestTenant(tenantType models.TenantType) *models.Tenant {
	now := time.Now().UTC()
	tenantID := uuid.New().String()

	var domainPtr *string
	if tenantType == models.TenantTypeOrganization {
		domain := "test-" + tenantID[:8] + ".example.com"
		domainPtr = &domain
	}

	return &models.Tenant{
		TenantID:  tenantID,
		Name:      "Test Organization",
		Domain:    domainPtr,
		Status:    models.TenantStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CreateTestTenants creates multiple test tenants
func CreateTestTenants(count int) []*models.Tenant {
	tenants := make([]*models.Tenant, count)

	for i := 0; i < count; i++ {
		tenantType := models.TenantTypeOrganization
		if i%2 == 0 {
			tenantType = models.TenantTypePersonal
		}

		tenant := CreateTestTenant(tenantType)
		tenant.Name = "Test Organization " + string(rune(i+1))
		tenants[i] = tenant
	}

	return tenants
}

// CreateTestUser creates a test user with realistic data
func CreateTestUser(tenantID string, withRole bool) *models.User {
	now := time.Now().UTC()
	userID := uuid.New().String()

	var roleID *int64
	if withRole {
		roleIDVal := int64(1 + (int(userID[0]) % 10)) // Deterministic role ID
		roleID = &roleIDVal
	}

	return &models.User{
		UserID:       userID,
		TenantID:     tenantID,
		RoleID:       roleID,
		Email:        "test-" + userID[:8] + "@example.com",
		PasswordHash: "$2a$12$hashedpasswordplaceholder",
		FullName:     "Test User",
		Status:       models.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  nil,
	}
}

// CreateTestUsers creates multiple test users for a tenant
func CreateTestUsers(count int, tenantID string) []*models.User {
	users := make([]*models.User, count)

	for i := 0; i < count; i++ {
		user := CreateTestUser(tenantID, i%2 == 0) // Alternate between with/without role
		user.FullName = "Test User " + string(rune(i+1))
		user.Email = "user" + string(rune(i+1)) + "-" + tenantID[:8] + "@example.com"
		users[i] = user
	}

	return users
}

// CreateTestSession creates a test session
func CreateTestSession(userID string) *models.Session {
	now := time.Now().UTC()
	sessionID := uuid.New().String()

	return &models.Session{
		SessionID: sessionID,
		UserID:    userID,
		DeviceID:  nil, // Set to nil to avoid foreign key constraint
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0 (Test Browser)",
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
}

// CreateTestSessions creates multiple test sessions for a user
func CreateTestSessions(count int, userID string) []*models.Session {
	sessions := make([]*models.Session, count)

	for i := 0; i < count; i++ {
		session := CreateTestSession(userID)
		session.SessionID = uuid.New().String()
		session.DeviceID = nil
		session.ExpiresAt = time.Now().UTC().Add(time.Duration(i+1) * time.Hour)
		sessions[i] = session
	}

	return sessions
}

// CreateTestRole creates a test role
func CreateTestRole(tenantID string) *models.Role {
	now := time.Now().UTC()
	roleID := int64(time.Now().UnixNano() % 1000000) // Deterministic role ID

	return &models.Role{
		RoleID:    roleID,
		TenantID:  tenantID,
		Name:      "Test Role " + uuid.New().String()[:8],
		CreatedAt: now,
	}
}

// CreateTestRoles creates multiple test roles for a tenant
func CreateTestRoles(count int, tenantID string) []*models.Role {
	roles := make([]*models.Role, count)

	for i := 0; i < count; i++ {
		role := CreateTestRole(tenantID)
		role.RoleID = int64(i + 1)
		role.Name = "Test Role " + string(rune(i+1))
		roles[i] = role
	}

	return roles
}

// CreateTestMFAMethod creates a test MFA method
func CreateTestMFAMethod(userID string, mfaType models.MFAMethodType) *models.UserMFAMethod {
	now := time.Now().UTC()
	mfaID := uuid.New().String()
	secret := "test-secret-" + mfaID[:8]

	return &models.UserMFAMethod{
		MFAID:     mfaID,
		UserID:    userID,
		Type:      mfaType,
		Secret:    &secret,
		Enabled:   false,
		CreatedAt: now,
	}
}

// CreateTestMFAMethods creates multiple test MFA methods for a user
func CreateTestMFAMethods(count int, userID string) []*models.UserMFAMethod {
	methods := make([]*models.UserMFAMethod, count)
	mfaTypes := []models.MFAMethodType{
		models.MFAMethodTypeTOTP,
		models.MFAMethodTypeSMS,
		models.MFAMethodTypeEmail,
	}

	for i := 0; i < count; i++ {
		methodType := mfaTypes[i%len(mfaTypes)]
		method := CreateTestMFAMethod(userID, methodType)
		method.MFAID = uuid.New().String()
		method.Enabled = i%2 == 0 // Alternate enabled/disabled
		methods[i] = method
	}

	return methods
}

// CreateTestDevice creates a test device
func CreateTestDevice(userID string) *models.Device {
	now := time.Now().UTC()
	deviceID := uuid.New().String()

	return &models.Device{
		DeviceID:   deviceID,
		UserID:     userID,
		DeviceName: "Test Device",
		LastSeenAt: now,
		CreatedAt:  now,
	}
}

// CreateTestDevices creates multiple test devices for a user
func CreateTestDevices(count int, userID string) []*models.Device {
	devices := make([]*models.Device, count)

	for i := 0; i < count; i++ {
		device := CreateTestDevice(userID)
		device.DeviceID = uuid.New().String()
		device.DeviceName = "Test Device " + string(rune(i+1))
		device.LastSeenAt = time.Now().UTC().Add(time.Duration(i) * time.Hour)
		devices[i] = device
	}

	return devices
}

// CreateTestInvitation creates a test invitation
func CreateTestInvitation(tenantID string, roleID int64) *models.Invitation {
	now := time.Now().UTC()
	invitationID := uuid.New().String()

	return &models.Invitation{
		InvitationID: invitationID,
		TenantID:     tenantID,
		RoleID:       roleID,
		Email:        "invite-" + invitationID[:8] + "@example.com",
		TokenHash:    "token-hash-" + invitationID[:8],
		Status:       models.InvitationStatusPending,
		ExpiresAt:    now.Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:    now,
	}
}

// CreateTestInvitations creates multiple test invitations for a tenant
func CreateTestInvitations(count int, tenantID string, roleID int64) []*models.Invitation {
	invitations := make([]*models.Invitation, count)

	for i := 0; i < count; i++ {
		invitation := CreateTestInvitation(tenantID, roleID)
		invitation.InvitationID = uuid.New().String()
		invitation.Email = "invite" + string(rune(i+1)) + "-" + tenantID[:8] + "@example.com"
		invitation.ExpiresAt = time.Now().UTC().Add(time.Duration(i+1) * 24 * time.Hour)
		invitations[i] = invitation
	}

	return invitations
}

// CreateTestPasswordReset creates a test password reset
func CreateTestPasswordReset(userID string) *models.PasswordReset {
	now := time.Now().UTC()
	resetID := uuid.New().String()

	return &models.PasswordReset{
		ResetID:   resetID,
		UserID:    userID,
		TokenHash: "reset-token-" + resetID[:8],
		Status:    models.PasswordResetStatusPending,
		ExpiresAt: now.Add(1 * time.Hour), // 1 hour
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CreateTestPasswordResets creates multiple test password resets for users
func CreateTestPasswordResets(userIDs []string) []*models.PasswordReset {
	resets := make([]*models.PasswordReset, len(userIDs))

	for i, userID := range userIDs {
		reset := CreateTestPasswordReset(userID)
		reset.ResetID = uuid.New().String()
		reset.ExpiresAt = time.Now().UTC().Add(time.Duration(i+1) * time.Hour)
		resets[i] = reset
	}

	return resets
}

// CreateTestRefreshToken creates a test refresh token
func CreateTestRefreshToken(userID string) *models.RefreshToken {
	now := time.Now().UTC()
	tokenID := uuid.New().String()

	return &models.RefreshToken{
		TokenID:   tokenID,
		UserID:    userID,
		TokenHash: "refresh-token-" + tokenID[:8],
		ExpiresAt: now.Add(30 * 24 * time.Hour), // 30 days
		CreatedAt: now,
	}
}

// CreateTestRefreshTokens creates multiple test refresh tokens for a user
func CreateTestRefreshTokens(count int, userID string) []*models.RefreshToken {
	tokens := make([]*models.RefreshToken, count)

	for i := 0; i < count; i++ {
		token := CreateTestRefreshToken(userID)
		token.TokenID = uuid.New().String()
		token.ExpiresAt = time.Now().UTC().Add(time.Duration(i+1) * 24 * time.Hour)
		tokens[i] = token
	}

	return tokens
}

// CreateTestAuditLog creates a test audit log
func CreateTestAuditLog(tenantID, userID, action string) *models.AuditLog {
	now := time.Now().UTC()
	logID := uuid.New().String()

	return &models.AuditLog{
		LogID:     logID,
		TenantID:  tenantID,
		UserID:    &userID,
		Action:    action,
		TargetID:  &userID,
		Metadata:  map[string]interface{}{"test": true},
		IPAddress: "192.168.1.100",
		UserAgent: "Mozilla/5.0 (Test Browser)",
		TraceID:   "trace-" + logID[:8],
		CreatedAt: now,
	}
}

// CreateTestAuditLogs creates multiple test audit logs
func CreateTestAuditLogs(count int, tenantID, userID string) []*models.AuditLog {
	logs := make([]*models.AuditLog, count)
	actions := []string{"create", "update", "delete", "login", "logout"}

	for i := 0; i < count; i++ {
		action := actions[i%len(actions)]
		log := CreateTestAuditLog(tenantID, userID, action)
		log.LogID = uuid.New().String()
		log.CreatedAt = time.Now().UTC().Add(-time.Duration(i) * time.Hour) // Staggered times
		logs[i] = log
	}

	return logs
}
