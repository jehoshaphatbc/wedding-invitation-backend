#!/bin/bash
cat << 'INNER_EOF' > /tmp/user_service_patch.go
func (s *UserService) GetTrashedUsers(page, perPage int, search, sort, order string) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.userRepo.FindTrashedAll(page, perPage, search, sort, order)
}

func (s *UserService) RestoreUser(id uuid.UUID, ip, userAgent string) error {
	err := s.userRepo.Restore(id)
	if err == nil {
		s.auditLog(nil, "user.restored", "users", &id, ip, userAgent)
	}
	return err
}

func (s *UserService) ForceDeleteUser(id uuid.UUID, ip, userAgent string) error {
	err := s.userRepo.ForceDelete(id)
	if err == nil {
		s.auditLog(nil, "user.force_deleted", "users", &id, ip, userAgent)
	}
	return err
}
INNER_EOF
cat /tmp/user_service_patch.go >> pkg/services/user_service.go
