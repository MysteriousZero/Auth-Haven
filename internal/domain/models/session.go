package models

import (
	"time"
)

type Session struct {
	SessionID string    `json:"session_id" db:"session_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	DeviceID  *string   `json:"device_id" db:"device_id"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
}

type Device struct {
	DeviceID   string    `json:"device_id" db:"device_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	DeviceName string    `json:"device_name" db:"device_name"`
	LastSeenAt time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type RefreshToken struct {
	TokenID    string    `json:"token_id" db:"token_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	TokenHash  string    `json:"-" db:"token_hash"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	Revoked    bool      `json:"revoked" db:"revoked"`
	ExpiresAt  time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
