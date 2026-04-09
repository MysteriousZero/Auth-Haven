-- Create tenants table
CREATE TABLE tenants (
    tenant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE, -- Only for org tenants
    type SMALLINT NOT NULL CHECK (type IN (1, 2)), -- 1: Organization, 2: Personal
    status SMALLINT NOT NULL DEFAULT 1 CHECK (status IN (1, 2, 3)), -- 1: Active, 2: Suspended, 3: Deleted
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create index on domain for org tenants
CREATE INDEX idx_tenants_domain ON tenants(domain) WHERE domain IS NOT NULL;

-- Create index on tenant status
CREATE INDEX idx_tenants_status ON tenants(status);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tenants_updated_at 
    BEFORE UPDATE ON tenants 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
