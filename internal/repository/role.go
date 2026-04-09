package repository

import (
	"auth-haven/internal/domain/errors"
	"auth-haven/internal/domain/interfaces"
	"auth-haven/internal/domain/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) interfaces.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) CreateRole(ctx context.Context, role *models.Role) error {
	query := `
		INSERT INTO roles (role_id, tenant_id, name, created_at)
		VALUES ($1, $2, $3, $4)
	`

	role.CreatedAt = time.Now().UTC()

	_, err := r.db.ExecContext(ctx, query,
		role.RoleID,
		role.TenantID,
		role.Name,
		role.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

func (r *roleRepository) GetRoleByID(ctx context.Context, roleID int64) (*models.Role, error) {
	query := `
		SELECT role_id, tenant_id, name, created_at
		FROM roles
		WHERE role_id = $1
	`

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.RoleID,
		&role.TenantID,
		&role.Name,
		&role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return &role, nil
}

func (r *roleRepository) ListRoles(ctx context.Context, tenantID string, limit, offset int) ([]models.Role, error) {
	query := `
		SELECT role_id, tenant_id, name, created_at
		FROM roles
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(
			&role.RoleID,
			&role.TenantID,
			&role.Name,
			&role.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *roleRepository) UpdateRole(ctx context.Context, role *models.Role) error {
	query := `
		UPDATE roles
		SET name = $2
		WHERE role_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, role.RoleID, role.Name)
	if err != nil {
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrRoleNotFound
	}

	return nil
}

func (r *roleRepository) DeleteRole(ctx context.Context, roleID int64) error {
	// First check if role is in use
	checkQuery := `SELECT COUNT(*) FROM users WHERE role_id = $1`
	var count int
	err := r.db.QueryRowContext(ctx, checkQuery, roleID).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if count > 0 {
		return errors.ErrRoleInUse
	}

	// Delete role permissions first
	deletePermsQuery := `DELETE FROM role_permissions WHERE role_id = $1`
	_, err = r.db.ExecContext(ctx, deletePermsQuery, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	// Delete role
	query := `DELETE FROM roles WHERE role_id = $1`

	result, err := r.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrRoleNotFound
	}

	return nil
}

func (r *roleRepository) CreateRolePermission(ctx context.Context, roleID int64, permissionKey string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_key)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionKey)
	if err != nil {
		return fmt.Errorf("failed to create role permission: %w", err)
	}

	return nil
}

func (r *roleRepository) DeleteRolePermissions(ctx context.Context, roleID int64) error {
	query := `DELETE FROM role_permissions WHERE role_id = $1`

	_, err := r.db.ExecContext(ctx, query, roleID)
	if err != nil {
		return fmt.Errorf("failed to delete role permissions: %w", err)
	}

	return nil
}

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID int64) ([]models.RolePermission, error) {
	query := `
		SELECT role_id, permission_key
		FROM role_permissions
		WHERE role_id = $1
		ORDER BY permission_key
	`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []models.RolePermission
	for rows.Next() {
		var permission models.RolePermission
		err := rows.Scan(
			&permission.RoleID,
			&permission.PermissionKey,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// Additional methods for role management
func (r *roleRepository) GetRoleByName(ctx context.Context, tenantID, name string) (*models.Role, error) {
	query := `
		SELECT role_id, tenant_id, name, created_at
		FROM roles
		WHERE tenant_id = $1 AND name = $2
	`

	var role models.Role
	err := r.db.QueryRowContext(ctx, query, tenantID, name).Scan(
		&role.RoleID,
		&role.TenantID,
		&role.Name,
		&role.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrRoleNotFound
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	return &role, nil
}

func (r *roleRepository) AddRolePermission(ctx context.Context, roleID int64, permissionKey string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_key)
		VALUES ($1, $2)
		ON CONFLICT (role_id, permission_key) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionKey)
	if err != nil {
		return fmt.Errorf("failed to add role permission: %w", err)
	}

	return nil
}

type deviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) interfaces.DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) CreateDevice(ctx context.Context, device *models.Device) error {
	query := `
		INSERT INTO devices (device_id, user_id, device_name, last_seen_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (device_id) DO UPDATE SET
			last_seen_at = EXCLUDED.last_seen_at,
			device_name = EXCLUDED.device_name
	`

	now := time.Now().UTC()
	device.CreatedAt = now
	device.LastSeenAt = now

	_, err := r.db.ExecContext(ctx, query,
		device.DeviceID,
		device.UserID,
		device.DeviceName,
		device.LastSeenAt,
		device.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

func (r *deviceRepository) GetDeviceByID(ctx context.Context, deviceID string) (*models.Device, error) {
	query := `
		SELECT device_id, user_id, device_name, last_seen_at, created_at
		FROM devices
		WHERE device_id = $1
	`

	var device models.Device
	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&device.DeviceID,
		&device.UserID,
		&device.DeviceName,
		&device.LastSeenAt,
		&device.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.ErrDeviceNotFound
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return &device, nil
}

func (r *deviceRepository) ListDevices(ctx context.Context, userID string, limit, offset int) ([]models.Device, error) {
	query := `
		SELECT device_id, user_id, device_name, last_seen_at, created_at
		FROM devices
		WHERE user_id = $1
		ORDER BY last_seen_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	var devices []models.Device
	for rows.Next() {
		var device models.Device
		err := rows.Scan(
			&device.DeviceID,
			&device.UserID,
			&device.DeviceName,
			&device.LastSeenAt,
			&device.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *deviceRepository) UpdateDevice(ctx context.Context, device *models.Device) error {
	query := `
		UPDATE devices
		SET device_name = $2, last_seen_at = $3
		WHERE device_id = $1
	`

	device.LastSeenAt = time.Now().UTC()

	result, err := r.db.ExecContext(ctx, query,
		device.DeviceID,
		device.DeviceName,
		device.LastSeenAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrDeviceNotFound
	}

	return nil
}

func (r *deviceRepository) DeleteDevice(ctx context.Context, deviceID string) error {
	query := `DELETE FROM devices WHERE device_id = $1`

	result, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrDeviceNotFound
	}

	return nil
}

func (r *deviceRepository) UpdateDeviceLastSeen(ctx context.Context, deviceID string) error {
	query := `UPDATE devices SET last_seen_at = NOW() WHERE device_id = $1`

	result, err := r.db.ExecContext(ctx, query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update device last seen: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrDeviceNotFound
	}

	return nil
}

