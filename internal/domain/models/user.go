package models

import (
	"time"
)

type UserStatus int16

const (
	UserStatusActive   UserStatus = 1
	UserStatusPending UserStatus = 2
	UserStatusDisabled UserStatus = 3
)

type User struct {
	UserID      string     `json:"user_id" db:"user_id"`
	TenantID    string     `json:"tenant_id" db:"tenant_id"`
	RoleID      *int64     `json:"role_id" db:"role_id"`
	Email       string     `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FullName    string     `json:"full_name" db:"full_name"`
	Status      UserStatus `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at" db:"last_login_at"`
}

func (u User) IsActive() bool {
	return u.Status == UserStatusActive
}

func (u User) IsPending() bool {
	return u.Status == UserStatusPending
}

func (u User) IsDisabled() bool {
	return u.Status == UserStatusDisabled
}

func (u User) HasRole() bool {
	return u.RoleID != nil
}

func (u User) IsPersonalUser() bool {
	return u.RoleID == nil
}
