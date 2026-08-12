package service

import (
	"auth-haven/internal/domain/interfaces"
	"fmt"
	"net/mail"
	"net/smtp"
	"net/url"
	"strings"
)

type emailProvider struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	baseURL  string
}

func NewEmailProvider(smtpHost string, smtpPort int, username, password, from string, baseURL ...string) interfaces.EmailProvider {
	publicURL := "http://localhost:3000"
	if len(baseURL) > 0 && baseURL[0] != "" {
		publicURL = strings.TrimRight(baseURL[0], "/")
	}
	return &emailProvider{smtpHost: smtpHost, smtpPort: smtpPort, username: username, password: password, from: from, baseURL: publicURL}
}

func (e *emailProvider) SendResetEmail(email, token string) error {
	link := e.baseURL + "/reset-password?token=" + url.QueryEscape(token)
	return e.send(email, "Auth Haven - Password Reset Request", "Use this link to reset your password:\n\n"+link+"\n\nIf you did not request this, ignore this email.")
}

func (e *emailProvider) SendInvitationEmail(email, token string) error {
	link := e.baseURL + "/register/invitation?token=" + url.QueryEscape(token)
	return e.send(email, "Auth Haven - Invitation to Join", "Use this link to accept your invitation:\n\n"+link)
}

func (e *emailProvider) send(to, subject, body string) error {
	if e.smtpHost == "" || e.smtpPort <= 0 || e.from == "" {
		return fmt.Errorf("email provider is not configured")
	}
	recipient, err := parseMailbox("recipient", to)
	if err != nil {
		return err
	}
	sender, err := parseMailbox("sender", e.from)
	if err != nil {
		return err
	}
	var auth smtp.Auth
	if e.username != "" {
		auth = smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	}
	// Keep the untrusted envelope recipient out of message headers entirely.
	message := []byte("From: " + sender.String() + "\r\nTo: undisclosed-recipients:;\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
	if err := smtp.SendMail(fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort), auth, sender.Address, []string{recipient.Address}, message); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func parseMailbox(field, value string) (*mail.Address, error) {
	if strings.ContainsAny(value, "\r\n") {
		return nil, fmt.Errorf("invalid email %s", field)
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Address == "" {
		return nil, fmt.Errorf("invalid email %s", field)
	}
	return address, nil
}
