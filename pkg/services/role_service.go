package services

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
)

type RoleService struct {
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
	auditRepo      repositories.AuditLogRepository
}

func NewRoleService(
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	auditRepo repositories.AuditLogRepository,
) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		auditRepo:      auditRepo,
	}
}

func (s *RoleService) GetAllRoles(page, perPage int, search, sort, order string) ([]models.Role, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.roleRepo.FindAll(page, perPage, search, sort, order)
}

func (s *RoleService) GetRoleByID(id uuid.UUID) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("role not found")
		}
		return nil, err
	}
	return role, nil
}

func (s *RoleService) CreateRole(req models.CreateRoleRequest, ip, userAgent string) (*models.Role, error) {
	_, findErr := s.roleRepo.FindByName(req.Name)
	if findErr == nil {
		return nil, errors.New("role name already exists")
	}

	role := &models.Role{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		IsSystem:    false,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return nil, err
	}

	if len(req.PermissionIDs) > 0 {
		s.roleRepo.AssignPermissions(role.ID, req.PermissionIDs)
	}

	s.auditLog(nil, "role.created", "roles", &role.ID, ip, userAgent)
	return s.roleRepo.FindByID(role.ID)
}

func (s *RoleService) UpdateRole(id uuid.UUID, req models.UpdateRoleRequest, ip, userAgent string) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	if role.Name == "super_admin" && req.Name != nil && *req.Name != "super_admin" {
		return nil, errors.New("cannot change super_admin role name")
	}

	if req.Name != nil && *req.Name != "" && role.Name != "super_admin" {
		role.Name = *req.Name
	}
	if req.DisplayName != nil {
		role.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		role.Description = req.Description
	}

	if err := s.roleRepo.Update(role); err != nil {
		return nil, err
	}

	if req.PermissionIDs != nil {
		if err := s.roleRepo.AssignPermissions(id, req.PermissionIDs); err != nil {
			return nil, err
		}
	}

	s.auditLog(nil, "role.updated", "roles", &role.ID, ip, userAgent)
	return s.roleRepo.FindByID(id)
}

func (s *RoleService) DeleteRole(id uuid.UUID, ip, userAgent string) error {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return errors.New("role not found")
	}

	if role.Name == "super_admin" {
		return errors.New("cannot delete super_admin role")
	}

	if err := s.roleRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "role.deleted", "roles", &role.ID, ip, userAgent)
	return nil
}

func (s *RoleService) AssignPermissions(id uuid.UUID, req models.AssignPermissionsRequest, ip, userAgent string) (*models.Role, error) {
	role, err := s.roleRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	if err := s.roleRepo.AssignPermissions(id, req.PermissionIDs); err != nil {
		return nil, err
	}

	s.auditLog(nil, "role.permissions_updated", "roles", &role.ID, ip, userAgent)
	return s.roleRepo.FindByID(id)
}

func (s *RoleService) GetAllPermissions() ([]models.Permission, error) {
	return s.permissionRepo.FindAll()
}

func (s *RoleService) CreatePermission(req models.CreatePermissionRequest, ip, userAgent string) (*models.Permission, error) {
	perm := &models.Permission{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
	}

	if err := s.permissionRepo.Create(perm); err != nil {
		return nil, err
	}

	s.auditLog(nil, "permission.created", "permissions", &perm.ID, ip, userAgent)
	return perm, nil
}

func (s *RoleService) UpdatePermission(id uuid.UUID, req models.UpdatePermissionRequest, ip, userAgent string) (*models.Permission, error) {
	perm, err := s.permissionRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("permission not found")
	}

	if req.DisplayName != nil {
		perm.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		perm.Description = req.Description
	}

	if err := s.permissionRepo.Update(perm); err != nil {
		return nil, err
	}

	s.auditLog(nil, "permission.updated", "permissions", &perm.ID, ip, userAgent)
	return perm, nil
}

func (s *RoleService) DeletePermission(id uuid.UUID, ip, userAgent string) error {
	perm, err := s.permissionRepo.FindByID(id)
	if err != nil {
		return errors.New("permission not found")
	}

	if err := s.permissionRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "permission.deleted", "permissions", &perm.ID, ip, userAgent)
	return nil
}

func (s *RoleService) GetRoleCount() (int64, error) {
	return s.roleRepo.Count()
}

func (s *RoleService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
	s.auditRepo.Create(&models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
	})
}
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
