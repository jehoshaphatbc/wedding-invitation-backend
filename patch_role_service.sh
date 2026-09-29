#!/bin/bash
cat << 'INNER_EOF' > /tmp/role_service_patch.go
func (s *RoleService) GetTrashedRoles(page, perPage int, search, sort, order string) ([]models.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.roleRepo.FindTrashedAll(page, perPage, search, sort, order)
}

func (s *RoleService) RestoreRole(id uuid.UUID, ip, userAgent string) error {
	err := s.roleRepo.Restore(id)
	if err == nil {
		s.auditLog(nil, "role.restored", "roles", &id, ip, userAgent)
	}
	return err
}

func (s *RoleService) ForceDeleteRole(id uuid.UUID, ip, userAgent string) error {
	err := s.roleRepo.ForceDelete(id)
	if err == nil {
		s.auditLog(nil, "role.force_deleted", "roles", &id, ip, userAgent)
	}
	return err
}
INNER_EOF
cat /tmp/role_service_patch.go >> pkg/services/role_service.go
