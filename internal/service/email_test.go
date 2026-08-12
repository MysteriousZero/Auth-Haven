package service

import "testing"

func TestEmailProviderRejectsMissingConfiguration(t *testing.T) {
	provider := NewEmailProvider("", 0, "", "", "")
	if err := provider.SendResetEmail("user@example.com", "secret-token"); err == nil {
		t.Fatal("SendResetEmail() accepted missing SMTP configuration")
	}
	if err := provider.SendInvitationEmail("user@example.com", "secret-token"); err == nil {
		t.Fatal("SendInvitationEmail() accepted missing SMTP configuration")
	}
}
