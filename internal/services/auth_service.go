package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/repositories"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/auth"
)

type AuthService struct {
	userRepo              repositories.UserRepository
	refreshTokenRepo      repositories.RefreshTokenRepository
	passwordResetRepo     repositories.PasswordResetTokenRepository
	emailVerificationRepo repositories.EmailVerificationTokenRepository
	roleRepo              repositories.RoleRepository
	profileRepo           repositories.ProfileRepository
	auditRepo             repositories.AuditLogRepository
	jwtManager            *auth.JWTManager
	cfg                   *config.Config
}

func NewAuthService(
	userRepo repositories.UserRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	passwordResetRepo repositories.PasswordResetTokenRepository,
	emailVerificationRepo repositories.EmailVerificationTokenRepository,
	roleRepo repositories.RoleRepository,
	profileRepo repositories.ProfileRepository,
	auditRepo repositories.AuditLogRepository,
	jwtManager *auth.JWTManager,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:              userRepo,
		refreshTokenRepo:      refreshTokenRepo,
		passwordResetRepo:     passwordResetRepo,
		emailVerificationRepo: emailVerificationRepo,
		roleRepo:              roleRepo,
		profileRepo:           profileRepo,
		auditRepo:             auditRepo,
		jwtManager:            jwtManager,
		cfg:                   cfg,
	}
}

func (s *AuthService) Register(req models.RegisterRequest, ip, userAgent string) (*models.AuthResponse, error) {
	_, err := s.userRepo.FindByEmail(strings.ToLower(req.Email))
	if err == nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	phone := req.Phone
	user := &models.User{
		Name:         req.Name,
		Email:        strings.ToLower(req.Email),
		Phone:        &phone,
		PasswordHash: hashedPassword,
		Status:       models.UserStatusActive,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	customerRole, err := s.roleRepo.FindByName("customer")
	if err == nil {
		s.userRepo.AssignRoles(user.ID, []uuid.UUID{customerRole.ID})
		user.Roles = []models.Role{*customerRole}
	}

	s.profileRepo.Create(&models.ClientProfile{
		UserID: user.ID,
	})

	s.auditLog(&user.ID, "user.created", "users", &user.ID, ip, userAgent)

	return s.generateTokenResponse(user, ip, userAgent)
}

func (s *AuthService) Login(req models.LoginRequest, ip, userAgent string) (*models.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(strings.ToLower(req.Email))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		s.auditLog(&user.ID, "login.failed", "users", &user.ID, ip, userAgent)
		return nil, errors.New("invalid email or password")
	}

	s.auditLog(&user.ID, "login.success", "users", &user.ID, ip, userAgent)
	s.userRepo.UpdateLastLogin(user.ID)

	return s.generateTokenResponse(user, ip, userAgent)
}

func (s *AuthService) RefreshToken(req models.RefreshTokenRequest, ip, userAgent string) (*models.TokenResponse, error) {
	tokenHash := auth.HashToken(req.RefreshToken)

	refreshToken, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if refreshToken.RevokedAt != nil {
		s.refreshTokenRepo.RevokeAllByUserID(refreshToken.UserID)
		return nil, errors.New("refresh token has been revoked, all sessions terminated")
	}

	user, err := s.userRepo.FindByID(refreshToken.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if user.Status != models.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	s.refreshTokenRepo.Revoke(refreshToken.ID)

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	newRefreshHash := auth.HashToken(newRefreshToken)

	expiry := s.jwtManager.GetRefreshExpiry()
	newRefreshModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: newRefreshHash,
		ExpiresAt: time.Now().Add(expiry),
		IPAddress: &ip,
		UserAgent: &userAgent,
	}
	s.refreshTokenRepo.Create(newRefreshModel)

	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.cfg.JWTAccessExpiry * 60,
	}, nil
}

func (s *AuthService) Logout(userID uuid.UUID, refreshTokenStr string, ip, userAgent string) error {
	tokenHash := auth.HashToken(refreshTokenStr)

	token, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	s.refreshTokenRepo.Revoke(token.ID)
	s.auditLog(&userID, "logout", "users", &userID, ip, userAgent)
	return nil
}

func (s *AuthService) LogoutAll(userID uuid.UUID, ip, userAgent string) error {
	s.refreshTokenRepo.RevokeAllByUserID(userID)
	s.auditLog(&userID, "logout.all", "users", &userID, ip, userAgent)
	return nil
}

func (s *AuthService) ForgotPassword(req models.ForgotPasswordRequest, ip, userAgent string) error {
	user, err := s.userRepo.FindByEmail(strings.ToLower(req.Email))
	if err != nil {
		return nil
	}

	resetToken, err := auth.GenerateRandomToken()
	if err != nil {
		return err
	}

	tokenHash := auth.HashToken(resetToken)

	s.passwordResetRepo.Create(&models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	})

	s.auditLog(&user.ID, "password.forgot_request", "users", &user.ID, ip, userAgent)
	return nil
}

func (s *AuthService) ResetPassword(req models.ResetPasswordRequest, ip, userAgent string) error {
	tokenHash := auth.HashToken(req.Token)

	resetToken, err := s.passwordResetRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user, err := s.userRepo.FindByID(resetToken.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	user.PasswordHash = hashedPassword
	s.userRepo.Update(user)
	s.passwordResetRepo.MarkAsUsed(resetToken.ID)
	s.refreshTokenRepo.RevokeAllByUserID(user.ID)

	s.auditLog(&user.ID, "password.reset", "users", &user.ID, ip, userAgent)
	return nil
}

func (s *AuthService) ChangePassword(userID uuid.UUID, req models.ChangePasswordRequest, ip, userAgent string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	user.PasswordHash = hashedPassword
	s.userRepo.Update(user)
	s.refreshTokenRepo.RevokeAllByUserID(userID)

	s.auditLog(&userID, "password.changed", "users", &userID, ip, userAgent)
	return nil
}

func (s *AuthService) VerifyEmail(token string, ip, userAgent string) error {
	tokenHash := auth.HashToken(token)

	verificationToken, err := s.emailVerificationRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return errors.New("invalid or expired verification token")
	}

	user, err := s.userRepo.FindByID(verificationToken.UserID)
	if err != nil {
		return errors.New("user not found")
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	user.Status = models.UserStatusActive
	s.userRepo.Update(user)
	s.emailVerificationRepo.MarkAsVerified(verificationToken.ID)

	s.auditLog(&user.ID, "email.verified", "users", &user.ID, ip, userAgent)
	return nil
}

func (s *AuthService) generateTokenResponse(user *models.User, ip, userAgent string) (*models.AuthResponse, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshHash := auth.HashToken(refreshToken)

	expiry := s.jwtManager.GetRefreshExpiry()
	refreshModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: refreshHash,
		ExpiresAt: time.Now().Add(expiry),
		IPAddress: &ip,
		UserAgent: &userAgent,
	}
	s.refreshTokenRepo.Create(refreshModel)

	return &models.AuthResponse{
		User:         models.ToUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    s.cfg.JWTAccessExpiry * 60,
	}, nil
}

func (s *AuthService) auditLog(userID *uuid.UUID, action, resourceType string, resourceID *uuid.UUID, ip, userAgent string) {
	s.auditRepo.Create(&models.AuditLog{
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    &ip,
		UserAgent:    &userAgent,
	})
}
