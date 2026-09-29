package services

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/auth"
)

type UserService struct {
	userRepo         repositories.UserRepository
	roleRepo         repositories.RoleRepository
	profileRepo      repositories.ProfileRepository
	auditRepo        repositories.AuditLogRepository
	refreshTokenRepo repositories.RefreshTokenRepository
}

func NewUserService(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	profileRepo repositories.ProfileRepository,
	auditRepo repositories.AuditLogRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		roleRepo:         roleRepo,
		profileRepo:      profileRepo,
		auditRepo:        auditRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (s *UserService) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) UpdateProfile(userID uuid.UUID, req models.UpdateProfileRequest, ip, userAgent string) (*models.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	s.auditLog(&userID, "user.updated", "users", &userID, ip, userAgent)
	return user, nil
}

func (s *UserService) GetAllUsers(page, perPage int, search, status, role, sort, order string) ([]models.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	return s.userRepo.FindAll(page, perPage, search, status, role, sort, order)
}

func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *UserService) CreateUser(req models.CreateUserRequest, ip, userAgent string) (*models.User, error) {
	_, findErr := s.userRepo.FindByEmail(strings.ToLower(req.Email))
	if findErr == nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:         req.Name,
		Email:        strings.ToLower(req.Email),
		Phone:        req.Phone,
		PasswordHash: hashedPassword,
		Status:       req.Status,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	if len(req.RoleIDs) > 0 {
		s.userRepo.AssignRoles(user.ID, req.RoleIDs)
	}

	s.profileRepo.Create(&models.ClientProfile{
		UserID: user.ID,
	})

	s.auditLog(nil, "user.created", "users", &user.ID, ip, userAgent)

	return s.userRepo.FindByID(user.ID)
}

func (s *UserService) UpdateUser(id uuid.UUID, req models.UpdateUserRequest, ip, userAgent string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	s.auditLog(nil, "user.updated", "users", &user.ID, ip, userAgent)
	return s.userRepo.FindByID(id)
}

func (s *UserService) AdminChangePassword(id uuid.UUID, req models.AdminChangePasswordRequest, ip, userAgent string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.refreshTokenRepo.RevokeAllByUserID(id)

	s.auditLog(nil, "user.password_forced_changed", "users", &user.ID, ip, userAgent)
	return nil
}

func (s *UserService) DeleteUser(id uuid.UUID, ip, userAgent string) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.userRepo.Delete(id); err != nil {
		return err
	}

	s.auditLog(nil, "user.deleted", "users", &user.ID, ip, userAgent)
	return nil
}

func (s *UserService) AssignRoles(id uuid.UUID, req models.AssignRolesRequest, ip, userAgent string) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if err := s.userRepo.AssignRoles(id, req.RoleIDs); err != nil {
		return nil, err
	}

	s.auditLog(nil, "user.roles_assigned", "users", &user.ID, ip, userAgent)
	return s.userRepo.FindByID(id)
}

func (s *UserService) GetUserCount() (int64, error) {
	return s.userRepo.Count()
}

func (s *UserService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
	s.auditRepo.Create(&models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
	})
}
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
