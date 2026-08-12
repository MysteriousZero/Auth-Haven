package service

import (
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
)

var errDelivery = errors.New("delivery unavailable")

type failingEmailProvider struct{}

func (failingEmailProvider) SendResetEmail(string, string) error      { return errDelivery }
func (failingEmailProvider) SendInvitationEmail(string, string) error { return errDelivery }

type deliveryUserRepository struct {
	interfaces.UserRepository
	user *models.User
}

func (r deliveryUserRepository) GetUserByEmail(context.Context, string, string) (*models.User, error) {
	if r.user == nil {
		return nil, errors.New("not found")
	}
	return r.user, nil
}
func (r deliveryUserRepository) GetUserByID(context.Context, string) (*models.User, error) {
	return r.user, nil
}

type deliveryPasswordResetRepository struct {
	interfaces.PasswordResetRepository
	created bool
}

func (r *deliveryPasswordResetRepository) CreatePasswordReset(context.Context, *models.PasswordReset) error {
	r.created = true
	return nil
}

type deliveryTokenService struct{ interfaces.TokenService }

func (deliveryTokenService) GenerateResetToken() string      { return "reset-token" }
func (deliveryTokenService) GenerateInvitationToken() string { return "invitation-token" }

type deliveryHasher struct{ interfaces.Hasher }

func (deliveryHasher) HashToken(token string) string { return "hashed-" + token }

type deliveryAuditRepository struct {
	interfaces.AuditRepository
	actions []string
}

func (r *deliveryAuditRepository) CreateAuditLog(_ context.Context, entry *models.AuditLog) error {
	r.actions = append(r.actions, entry.Action)
	return nil
}

func TestPasswordResetDeliveryFailureIsObservableWithoutEnumeration(t *testing.T) {
	user := &models.User{UserID: "user", TenantID: "tenant", Status: models.UserStatusActive}
	resets := &deliveryPasswordResetRepository{}
	audits := &deliveryAuditRepository{}
	service := NewPasswordService(deliveryUserRepository{user: user}, resets, nil, deliveryHasher{}, deliveryTokenService{}, audits, failingEmailProvider{})
	var logs bytes.Buffer
	previous := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(previous) })

	if err := service.RequestPasswordReset(context.Background(), "tenant", "user@example.com"); err != nil {
		t.Fatalf("RequestPasswordReset() exposed delivery failure: %v", err)
	}
	if !resets.created {
		t.Fatal("password reset was not persisted")
	}
	if !strings.Contains(logs.String(), "password reset delivery failed") {
		t.Fatalf("delivery failure was not logged: %q", logs.String())
	}
	if strings.Contains(logs.String(), "reset-token") {
		t.Fatal("raw reset token was logged")
	}
}

type deliveryTenantRepository struct {
	interfaces.TenantRepository
	created bool
}

func (r *deliveryTenantRepository) GetTenantByID(context.Context, string) (*models.Tenant, error) {
	return &models.Tenant{TenantID: "tenant", Type: models.TenantTypeOrganization}, nil
}
func (r *deliveryTenantRepository) CreateInvitation(context.Context, *models.Invitation) error {
	r.created = true
	return nil
}

type deliveryRoleRepository struct{ interfaces.RoleRepository }

func (deliveryRoleRepository) GetRoleByID(context.Context, int64) (*models.Role, error) {
	return &models.Role{RoleID: 2, TenantID: "tenant"}, nil
}
func (deliveryRoleRepository) GetRolePermissions(context.Context, int64) ([]models.RolePermission, error) {
	return []models.RolePermission{{RoleID: 1, PermissionKey: "users.write"}}, nil
}

type invitationUserRepository struct {
	interfaces.UserRepository
	actor *models.User
}

func (r invitationUserRepository) GetUserByID(context.Context, string) (*models.User, error) {
	return r.actor, nil
}
func (invitationUserRepository) GetUserByEmail(context.Context, string, string) (*models.User, error) {
	return nil, errors.New("not found")
}

func TestInvitationDeliveryFailureIsReturnedAfterPersistenceAndAudit(t *testing.T) {
	roleID := int64(1)
	users := invitationUserRepository{actor: &models.User{UserID: "actor", TenantID: "tenant", RoleID: &roleID, Status: models.UserStatusActive}}
	tenants := &deliveryTenantRepository{}
	audits := &deliveryAuditRepository{}
	service := NewInvitationService(tenants, users, deliveryRoleRepository{}, audits, failingEmailProvider{}, deliveryTokenService{}, deliveryHasher{})

	invitation, err := service.SendInvitation(context.Background(), "actor", "tenant", "invitee@example.com", 2)
	if err == nil || !strings.Contains(err.Error(), "delivery failed") {
		t.Fatalf("SendInvitation() error = %v", err)
	}
	if invitation != nil {
		t.Fatal("failed delivery returned an invitation as successful")
	}
	if !tenants.created {
		t.Fatal("invitation was not persisted before delivery")
	}
	if len(audits.actions) != 1 || audits.actions[0] != models.AuditInvitationSent {
		t.Fatalf("audit actions = %v", audits.actions)
	}
}
