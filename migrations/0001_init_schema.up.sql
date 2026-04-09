CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Tenants
CREATE TABLE tenants (
    tenant_id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(255) NOT NULL,
    domain      VARCHAR(255) UNIQUE,
    status      SMALLINT NOT NULL DEFAULT 1, -- 1=active, 2=suspended, 3=deleted
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_name ON tenants(name);

-- 3. Roles
CREATE TABLE roles (
    role_id     BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, name)
);

-- 4. Role Permissions
CREATE TABLE role_permissions (
    role_id         BIGINT NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    permission_key  VARCHAR(255) NOT NULL,
    PRIMARY KEY (role_id, permission_key)
);

-- 2. Users
CREATE TABLE users (
    user_id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    role_id       BIGINT REFERENCES roles(role_id) ON DELETE SET NULL,
    email         VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255) NOT NULL,
    status        SMALLINT NOT NULL DEFAULT 1, -- 1=active, 2=pending, 3=disabled
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMP,
    UNIQUE (tenant_id, email)
);

CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_users_status ON users(status);

-- 5. Refresh Tokens
CREATE TABLE refresh_tokens (
    token_id    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    user_agent  TEXT,
    ip_address  VARCHAR(45),
    revoked     BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at  TIMESTAMP NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- 7. Devices
CREATE TABLE devices (
    device_id    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    device_name  VARCHAR(255) NOT NULL,
    last_seen_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 6. Sessions
CREATE TABLE sessions (
    session_id  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    device_id   UUID REFERENCES devices(device_id) ON DELETE CASCADE,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMP NOT NULL
);

CREATE INDEX idx_sessions_user_id_expires_at ON sessions(user_id, expires_at);

-- 8. User MFA Methods
CREATE TABLE user_mfa_methods (
    mfa_id       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id      UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type         SMALLINT NOT NULL, -- 1=TOTP, 2=SMS
    secret       VARCHAR(255), -- TOTP secret
    phone_number VARCHAR(50),  -- required for SMS
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 9. Password Resets
CREATE TABLE password_resets (
    reset_id   UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    status     SMALLINT NOT NULL DEFAULT 1, -- 1=pending, 2=used, 3=expired
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 10. Invitations
CREATE TABLE invitations (
    invitation_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    role_id       BIGINT NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    email         VARCHAR(255) NOT NULL,
    token_hash    VARCHAR(255) NOT NULL UNIQUE,
    status        SMALLINT NOT NULL DEFAULT 1, -- 1=pending, 2=accepted, 3=expired
    expires_at    TIMESTAMP NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 11. Audit Logs
CREATE TABLE audit_logs (
    log_id      UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID REFERENCES users(user_id) ON DELETE SET NULL,
    tenant_id   UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    action      VARCHAR(255) NOT NULL,
    ip_address  VARCHAR(45),
    user_agent  TEXT,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_id_created_at ON audit_logs(tenant_id, created_at);
CREATE INDEX idx_audit_logs_user_id_created_at ON audit_logs(user_id, created_at);