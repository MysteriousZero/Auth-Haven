package service

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	domainerrors "auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"

	"github.com/pquerna/otp/totp"
)

type mfaLifecycleRepo struct {
	interfaces.MFAMethodRepository
	methods map[string]models.UserMFAMethod
}

func (r *mfaLifecycleRepo) CreateMFAMethod(_ context.Context, method *models.UserMFAMethod) error {
	r.methods[method.MFAID] = *method
	return nil
}

func (r *mfaLifecycleRepo) ListMFAMethods(_ context.Context, userID string) ([]models.UserMFAMethod, error) {
	var result []models.UserMFAMethod
	for _, method := range r.methods {
		if method.UserID == userID {
			result = append(result, method)
		}
	}
	return result, nil
}

func (r *mfaLifecycleRepo) UpdateMFAMethod(_ context.Context, id string, enabled bool) error {
	method, ok := r.methods[id]
	if !ok {
		return domainerrors.ErrMFAMethodNotFound
	}
	method.Enabled = enabled
	r.methods[id] = method
	return nil
}

type fixedTOTPGenerator struct {
	interfaces.TOTPGenerator
	secret string
}

func (g fixedTOTPGenerator) GenerateSecret() string                  { return g.secret }
func (g fixedTOTPGenerator) GenerateQRCodeURL(string, string) string { return "otpauth://test" }

func TestMFATOTPLifecycleRejectsWrongOwnerAndInvalidCode(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	key := []byte("0123456789abcdef0123456789abcdef")
	user := &models.User{UserID: "user-a", TenantID: "tenant-a", Email: "person@example.com", Status: models.UserStatusActive}
	repo := &mfaLifecycleRepo{methods: make(map[string]models.UserMFAMethod)}
	users := userRepoStub{getByID: func(context.Context, string) (*models.User, error) { return user, nil }}
	service := NewMFAService(repo, users, auditRepoStub{}, fixedTOTPGenerator{secret: secret}, key)

	enrollment, err := service.EnrollTOTP(context.Background(), user.UserID)
	if err != nil || enrollment.Secret != secret || enrollment.MFAID == "" {
		t.Fatalf("EnrollTOTP() = %#v, %v", enrollment, err)
	}
	if err := service.ActivateTOTP(context.Background(), "attacker", enrollment.MFAID, "000000"); !stderrors.Is(err, domainerrors.ErrMFAMethodNotFound) {
		t.Fatalf("wrong-owner ActivateTOTP() error = %v", err)
	}
	if err := service.ActivateTOTP(context.Background(), user.UserID, enrollment.MFAID, "000000"); !stderrors.Is(err, domainerrors.ErrInvalidMFACode) {
		t.Fatalf("invalid-code ActivateTOTP() error = %v", err)
	}

	code, err := totp.GenerateCode(secret, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if err := service.ActivateTOTP(context.Background(), user.UserID, enrollment.MFAID, code); err != nil {
		t.Fatalf("ActivateTOTP() error = %v", err)
	}
	if !repo.methods[enrollment.MFAID].Enabled {
		t.Fatal("MFA method was not enabled")
	}
	if err := service.ActivateTOTP(context.Background(), user.UserID, enrollment.MFAID, code); !stderrors.Is(err, domainerrors.ErrMFAAlreadyEnabled) {
		t.Fatalf("replayed ActivateTOTP() error = %v", err)
	}

	methods, err := service.ListMFAMethods(context.Background(), user.UserID)
	if err != nil || len(methods) != 1 || methods[0].Secret == nil || *methods[0].Secret != "" {
		t.Fatalf("ListMFAMethods() = %#v, %v; secret must be redacted", methods, err)
	}
}
