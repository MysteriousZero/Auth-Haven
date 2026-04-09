package service

import (
	"auth-haven/internal/domain/interfaces"
	"fmt"
)

type emailProvider struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
}

func NewEmailProvider(smtpHost string, smtpPort int, username, password, from string) interfaces.EmailProvider {
	return &emailProvider{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		from:     from,
	}
}

func (e *emailProvider) SendResetEmail(email, token string) error {
	// In production, use a proper SMTP client like net/smtp or a service like SendGrid
	// For now, we'll just log the email that would be sent
	resetURL := fmt.Sprintf("https://api.auth-haven.com/v1/auth/reset-password?token=%s", token)
	
	subject := "Auth Haven - Password Reset Request"
	body := fmt.Sprintf(`
Hello,

You have requested to reset your password for your Auth Haven account.

Click the link below to reset your password:
%s

This link will expire in 1 hour.

If you did not request a password reset, please ignore this email.

Best regards,
The Auth Haven Team
`, resetURL)

	fmt.Printf("=== EMAIL TO BE SENT ===\n")
	fmt.Printf("To: %s\n", email)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Body:\n%s\n", body)
	fmt.Printf("========================\n")
	
	// TODO: Implement actual email sending
	// Example with net/smtp:
	/*
	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	msg := "To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body
	
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort),
		auth,
		e.from,
		[]string{email},
		[]byte(msg),
	)
	
	return err
	*/
	
	return nil
}

func (e *emailProvider) SendInvitationEmail(email, token string) error {
	// In production, use a proper SMTP client or email service
	// For now, we'll just log the email that would be sent
	invitationURL := fmt.Sprintf("https://api.auth-haven.com/v1/auth/register/invitation?token=%s", token)
	
	subject := "Auth Haven - Invitation to Join"
	body := fmt.Sprintf(`
Hello,

You have been invited to join Auth Haven!

Click the link below to complete your registration:
%s

This invitation link will expire in 7 days.

If you are not interested in this invitation, please ignore this email.

Best regards,
The Auth Haven Team
`, invitationURL)

	fmt.Printf("=== EMAIL TO BE SENT ===\n")
	fmt.Printf("To: %s\n", email)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Body:\n%s\n", body)
	fmt.Printf("========================\n")
	
	// TODO: Implement actual email sending
	return nil
}

