package models

import (
	"time"
)

type Role struct {
	RoleID    int64     `json:"role_id" db:"role_id"`
	TenantID  string    `json:"tenant_id" db:"tenant_id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type RolePermission struct {
	RoleID         int64  `json:"role_id" db:"role_id"`
	PermissionKey string `json:"permission_key" db:"permission_key"`
}

type CreateRoleRequest struct {
	TenantID string `json:"tenant_id"`
	Name     string `json:"name"`
}

type UpdateRoleRequest struct {
	Name            *string   `json:"name,omitempty"`
	PermissionKeys  []string  `json:"permission_keys,omitempty"`
}
