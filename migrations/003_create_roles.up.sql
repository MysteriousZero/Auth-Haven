-- Create roles table
CREATE TABLE roles (
    role_id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Unique role name per tenant
    UNIQUE(tenant_id, name)
);

-- Create role_permissions table
CREATE TABLE role_permissions (
    role_id BIGINT NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_key VARCHAR(255) NOT NULL,
    
    PRIMARY KEY (role_id, permission_key)
);

-- Create indexes
CREATE INDEX idx_roles_tenant_id ON roles(tenant_id);
CREATE INDEX idx_role_permissions_role_id ON role_permissions(role_id);
