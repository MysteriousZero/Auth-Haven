package models

import (
	"time"
)

type TenantType int16

const (
	TenantTypeOrganization TenantType = 1
	TenantTypePersonal     TenantType = 2
)

type TenantStatus int16

const (
	TenantStatusActive    TenantStatus = 1
	TenantStatusSuspended TenantStatus = 2
	TenantStatusDeleted   TenantStatus = 3
)

type Tenant struct {
	TenantID  string       `json:"tenant_id" db:"tenant_id"`
	Name      string       `json:"name" db:"name"`
	Domain    *string      `json:"domain" db:"domain"`
	Type      TenantType   `json:"type" db:"type"`
	Status    TenantStatus `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

func (t Tenant) IsOrganization() bool {
	return t.Type == TenantTypeOrganization
}

func (t Tenant) IsPersonal() bool {
	return t.Type == TenantTypePersonal
}

func (t Tenant) IsActive() bool {
	return t.Status == TenantStatusActive
}

func (t Tenant) IsSuspended() bool {
	return t.Status == TenantStatusSuspended
}

func (t Tenant) IsDeleted() bool {
	return t.Status == TenantStatusDeleted
}
