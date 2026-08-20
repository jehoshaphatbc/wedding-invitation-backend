DELETE FROM role_permissions WHERE role_id IN (SELECT id FROM roles WHERE is_system = true);
DELETE FROM roles WHERE is_system = true;
DELETE FROM permissions WHERE name IN (
    'profile.view', 'profile.update',
    'user.view', 'user.create', 'user.update', 'user.delete',
    'role.view', 'role.create', 'role.update', 'role.delete', 'role.assign',
    'permission.view',
    'invitation.view', 'invitation.create', 'invitation.update', 'invitation.delete', 'invitation.publish',
    'template.view', 'template.create', 'template.update', 'template.delete',
    'rsvp.view', 'rsvp.export',
    'contact.view', 'contact.delete'
);
