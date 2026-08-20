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

INSERT INTO roles (id, name, display_name, description, is_system, created_at, updated_at) VALUES
    (uuid_generate_v4(), 'super_admin', 'Super Admin', 'Full system access', true, NOW(), NOW()),
    (uuid_generate_v4(), 'admin', 'Admin', 'Manage users, content and system', true, NOW(), NOW()),
    (uuid_generate_v4(), 'customer', 'Customer', 'Regular user with invitation access', true, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, NOW()
FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT DO NOTHING;

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
