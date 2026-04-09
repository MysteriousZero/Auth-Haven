-- Create user_mfa_methods table
CREATE TABLE user_mfa_methods (
    mfa_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    type SMALLINT NOT NULL CHECK (type IN (1, 2, 3)), -- 1: TOTP, 2: SMS, 3: Email
    secret TEXT, -- Encrypted at rest, NULL for SMS/Email
    phone_number VARCHAR(20), -- Only for SMS
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Only one enabled method per type per user
    UNIQUE(user_id, type, enabled) WHERE enabled = TRUE
);

-- Create password_resets table
CREATE TABLE password_resets (
    reset_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    status SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (1, 2, 3)), -- 1: Pending, 2: Used, 3: Expired
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create invitations table
CREATE TABLE invitations (
    invitation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(tenant_id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(role_id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    status SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (1, 2, 3)), -- 1: Pending, 2: Accepted, 3: Expired
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Unique invitation per email per tenant
    UNIQUE(tenant_id, email, status) WHERE status = 1
);

-- Create indexes
CREATE INDEX idx_user_mfa_methods_user_id ON user_mfa_methods(user_id);
CREATE INDEX idx_user_mfa_methods_enabled ON user_mfa_methods(user_id, enabled) WHERE enabled = TRUE;
CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX idx_password_resets_hash ON password_resets(token_hash);
CREATE INDEX idx_password_resets_expires_at ON password_resets(expires_at);
CREATE INDEX idx_invitations_tenant_id ON invitations(tenant_id);
CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_hash ON invitations(token_hash);
