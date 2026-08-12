package service

import (
	"auth-haven/internal/domain/interfaces"
	"fmt"
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
	var auth smtp.Auth
	if e.username != "" {
		auth = smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	}
	message := []byte("From: " + e.from + "\r\nTo: " + to + "\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
	if err := smtp.SendMail(fmt.Sprintf("%s:%d", e.smtpHost, e.smtpPort), auth, e.from, []string{to}, message); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
