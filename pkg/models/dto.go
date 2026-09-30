package models

import (
	"time"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,max=150"`
	Email    string `json:"email" binding:"required,email"`
	Phone    string `json:"phone" binding:"omitempty,max=30"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type AdminChangePasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ChangeEmailRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type UpdateProfileRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=150"`
	Phone *string `json:"phone" binding:"omitempty,max=30"`
}

type CreateUserRequest struct {
	Name     string      `json:"name" binding:"required,max=150"`
	Email    string      `json:"email" binding:"required,email"`
	Phone    *string     `json:"phone" binding:"omitempty,max=30"`
	Password string      `json:"password" binding:"required,min=8"`
	Status   UserStatus  `json:"status" binding:"required,oneof=active inactive suspended pending"`
	RoleIDs  []uuid.UUID `json:"role_ids" binding:"required"`
}

type UpdateUserRequest struct {
	Name   *string     `json:"name" binding:"omitempty,max=150"`
	Phone  *string     `json:"phone" binding:"omitempty,max=30"`
	Status *UserStatus `json:"status" binding:"omitempty,oneof=active inactive suspended pending"`
}

type AssignRolesRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids" binding:"required"`
}

type CreateRoleRequest struct {
	Name          string      `json:"name" binding:"required,max=50"`
	DisplayName   string      `json:"display_name" binding:"required,max=100"`
	Description   *string     `json:"description"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type UpdateRoleRequest struct {
	Name          *string     `json:"name" binding:"omitempty,max=50"`
	DisplayName   *string     `json:"display_name" binding:"omitempty,max=100"`
	Description   *string     `json:"description"`
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids" binding:"required"`
}

type CreatePermissionRequest struct {
	Name        string  `json:"name" binding:"required,max=100"`
	DisplayName string  `json:"display_name" binding:"required,max=150"`
	Description *string `json:"description"`
}

type UpdatePermissionRequest struct {
	DisplayName *string `json:"display_name" binding:"omitempty,max=150"`
	Description *string `json:"description"`
}

type UserResponse struct {
	ID              uuid.UUID      `json:"id"`
	Name            string         `json:"name"`
	Email           string         `json:"email"`
	Phone           *string        `json:"phone,omitempty"`
	Status          UserStatus     `json:"status"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	Roles           []RoleResponse `json:"roles,omitempty"`
}

type RoleResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	DisplayName string               `json:"display_name"`
	Description *string              `json:"description,omitempty"`
	IsSystem    bool                 `json:"is_system"`
	Permissions []PermissionResponse `json:"permissions,omitempty"`
}

type PermissionResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func ToUserResponse(user *User) UserResponse {
	roles := make([]RoleResponse, 0)
	for _, role := range user.Roles {
		perms := make([]PermissionResponse, 0)
		for _, p := range role.Permissions {
			perms = append(perms, PermissionResponse{
				ID:          p.ID,
				Name:        p.Name,
				DisplayName: p.DisplayName,
			})
		}
		roles = append(roles, RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			DisplayName: role.DisplayName,
			Description: role.Description,
			IsSystem:    role.IsSystem,
			Permissions: perms,
		})
	}

	return UserResponse{
		ID:              user.ID,
		Name:            user.Name,
		Email:           user.Email,
		Phone:           user.Phone,
		Status:          user.Status,
		EmailVerifiedAt: user.EmailVerifiedAt,
		LastLoginAt:     user.LastLoginAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
		Roles:           roles,
	}
}

func ToRoleResponse(role *Role) RoleResponse {
	perms := make([]PermissionResponse, 0)
	for _, p := range role.Permissions {
		perms = append(perms, PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			DisplayName: p.DisplayName,
		})
	}

	return RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		DisplayName: role.DisplayName,
		Description: role.Description,
		IsSystem:    role.IsSystem,
		Permissions: perms,
	}
}

type UpdateCompanySettingRequest struct {
	Name        string `form:"name" binding:"required"`
	Address     string `form:"address"`
	Description string `form:"description"`
}

type CompanySettingResponse struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	Description   string `json:"description"`
	FaviconURL    string `json:"favicon_url"`
	LogoLongURL   string `json:"logo_long_url"`
	LogoSquareURL string `json:"logo_square_url"`
}

func ToCompanySettingResponse(setting *CompanySetting) CompanySettingResponse {
	return CompanySettingResponse{
		Name:          setting.Name,
		Address:       setting.Address,
		Description:   setting.Description,
		FaviconURL:    setting.FaviconURL,
		LogoLongURL:   setting.LogoLongURL,
		LogoSquareURL: setting.LogoSquareURL,
	}
}
