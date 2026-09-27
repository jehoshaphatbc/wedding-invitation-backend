-- ============================================
-- Wedding Invitation Backend - Full Schema
-- Untuk Neon PostgreSQL
-- ============================================

-- 1. UUID Extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 2. Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    phone VARCHAR(30),
    password_hash TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    email_verified_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 3. Roles Table
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);

-- 4. Permissions Table
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(150) NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);

-- 5. User Roles Junction
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role_id ON user_roles(role_id);

-- 6. Role Permissions Junction
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_role_permissions_role_id ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_permissions_permission_id ON role_permissions(permission_id);

-- 7. Client Profiles
CREATE TABLE IF NOT EXISTS client_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(150),
    address TEXT,
    city VARCHAR(100),
    province VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'Indonesia',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_client_profiles_user_id ON client_profiles(user_id);

-- 8. Refresh Tokens
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- 9. Password Reset Tokens
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);

-- 10. Email Verification Tokens
CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token_hash ON email_verification_tokens(token_hash);

-- 11. Audit Logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- ============================================
-- SEED DATA
-- ============================================

-- 12. Seed Permissions (25)
INSERT INTO permissions (id, name, display_name, created_at, updated_at) VALUES
    (uuid_generate_v4(), 'profile.view', 'View Profile', NOW(), NOW()),
    (uuid_generate_v4(), 'profile.update', 'Update Profile', NOW(), NOW()),
    (uuid_generate_v4(), 'user.view', 'View Users', NOW(), NOW()),
    (uuid_generate_v4(), 'user.create', 'Create Users', NOW(), NOW()),
    (uuid_generate_v4(), 'user.update', 'Update Users', NOW(), NOW()),
    (uuid_generate_v4(), 'user.delete', 'Delete Users', NOW(), NOW()),
    (uuid_generate_v4(), 'role.view', 'View Roles', NOW(), NOW()),
    (uuid_generate_v4(), 'role.create', 'Create Roles', NOW(), NOW()),
    (uuid_generate_v4(), 'role.update', 'Update Roles', NOW(), NOW()),
    (uuid_generate_v4(), 'role.delete', 'Delete Roles', NOW(), NOW()),
    (uuid_generate_v4(), 'role.assign', 'Assign Roles', NOW(), NOW()),
    (uuid_generate_v4(), 'permission.view', 'View Permissions', NOW(), NOW()),
    (uuid_generate_v4(), 'invitation.view', 'View Invitations', NOW(), NOW()),
    (uuid_generate_v4(), 'invitation.create', 'Create Invitations', NOW(), NOW()),
    (uuid_generate_v4(), 'invitation.update', 'Update Invitations', NOW(), NOW()),
    (uuid_generate_v4(), 'invitation.delete', 'Delete Invitations', NOW(), NOW()),
    (uuid_generate_v4(), 'invitation.publish', 'Publish Invitations', NOW(), NOW()),
    (uuid_generate_v4(), 'template.view', 'View Templates', NOW(), NOW()),
    (uuid_generate_v4(), 'template.create', 'Create Templates', NOW(), NOW()),
    (uuid_generate_v4(), 'template.update', 'Update Templates', NOW(), NOW()),
    (uuid_generate_v4(), 'template.delete', 'Delete Templates', NOW(), NOW()),
    (uuid_generate_v4(), 'rsvp.view', 'View RSVPs', NOW(), NOW()),
    (uuid_generate_v4(), 'rsvp.export', 'Export RSVPs', NOW(), NOW()),
    (uuid_generate_v4(), 'contact.view', 'View Contacts', NOW(), NOW()),
    (uuid_generate_v4(), 'contact.delete', 'Delete Contacts', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 13. Seed Roles (3)
INSERT INTO roles (id, name, display_name, description, is_system, created_at, updated_at) VALUES
    (uuid_generate_v4(), 'super_admin', 'Super Admin', 'Full system access', true, NOW(), NOW()),
    (uuid_generate_v4(), 'admin', 'Admin', 'Manage users, content and system', true, NOW(), NOW()),
    (uuid_generate_v4(), 'customer', 'Customer', 'Regular user with invitation access', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 14. Assign All Permissions to super_admin
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT DO NOTHING;

-- 15. Assign Permissions to admin
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r, permissions p
WHERE r.name = 'admin'
AND p.name IN (
    'profile.view', 'profile.update',
    'user.view', 'user.create', 'user.update',
    'role.view', 'role.assign',
    'permission.view',
    'invitation.view', 'invitation.create', 'invitation.update', 'invitation.delete', 'invitation.publish',
    'template.view', 'template.create', 'template.update', 'template.delete',
    'rsvp.view', 'rsvp.export',
    'contact.view'
)
ON CONFLICT DO NOTHING;

-- 16. Assign Permissions to customer
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r, permissions p
WHERE r.name = 'customer'
AND p.name IN (
    'profile.view', 'profile.update',
    'invitation.view', 'invitation.create', 'invitation.update', 'invitation.delete', 'invitation.publish',
    'template.view',
    'rsvp.view', 'rsvp.export'
)
ON CONFLICT DO NOTHING;

-- 17. Create Super Admin User
-- Password: SuperAdmin123!
INSERT INTO users (id, name, email, password_hash, status, created_at, updated_at)
VALUES (
    uuid_generate_v4(),
    'Super Admin',
    'admin@wedding.com',
    '$2a$10$HnFIWVFNEmQ5AZeVlvpO9.V9jIFMDOpdxdGg5qqpTL5AJbL.VWoO2',
    'active',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- 18. Assign super_admin Role to Super Admin User
INSERT INTO user_roles (user_id, role_id, created_at)
SELECT u.id, r.id, NOW()
FROM users u, roles r
WHERE u.email = 'admin@wedding.com'
AND r.name = 'super_admin'
ON CONFLICT DO NOTHING;

-- 19. Create Client Profile for Super Admin
INSERT INTO client_profiles (id, user_id, country, created_at, updated_at)
SELECT uuid_generate_v4(), u.id, 'Indonesia', NOW(), NOW()
FROM users u
WHERE u.email = 'admin@wedding.com'
AND NOT EXISTS (
    SELECT 1 FROM client_profiles cp WHERE cp.user_id = u.id
);

-- ============================================
-- DONE!
-- ============================================
