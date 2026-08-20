package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusSuspend  UserStatus = "suspended"
	UserStatusPending  UserStatus = "pending"
)

type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name            string         `gorm:"type:varchar(150);not null" json:"name"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Phone           *string        `gorm:"type:varchar(30)" json:"phone,omitempty"`
	PasswordHash    string         `gorm:"type:text;not null" json:"-"`
	Status          UserStatus     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	EmailVerifiedAt *time.Time     `gorm:"type:timestamptz" json:"email_verified_at,omitempty"`
	LastLoginAt     *time.Time     `gorm:"type:timestamptz" json:"last_login_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	Roles           []Role         `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	ClientProfile   *ClientProfile  `gorm:"foreignKey:UserID" json:"client_profile,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash  string     `gorm:"type:text;not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	RevokedAt  *time.Time `gorm:"type:timestamptz" json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `gorm:"type:timestamptz" json:"last_used_at,omitempty"`
	IPAddress  *string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent  *string    `gorm:"type:text" json:"user_agent,omitempty"`
}

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"type:text;not null" json:"-"`
	ExpiresAt time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"type:timestamptz" json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (p *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type EmailVerificationToken struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Email      string     `gorm:"type:varchar(255);not null" json:"email"`
	TokenHash  string     `gorm:"type:text;not null" json:"-"`
	ExpiresAt  time.Time  `gorm:"type:timestamptz;not null" json:"expires_at"`
	VerifiedAt *time.Time `gorm:"type:timestamptz" json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (e *EmailVerificationToken) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

type ClientProfile struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	CompanyName *string   `gorm:"type:varchar(150)" json:"company_name,omitempty"`
	Address     *string   `gorm:"type:text" json:"address,omitempty"`
	City        *string   `gorm:"type:varchar(100)" json:"city,omitempty"`
	Province    *string   `gorm:"type:varchar(100)" json:"province,omitempty"`
	PostalCode  *string   `gorm:"type:varchar(20)" json:"postal_code,omitempty"`
	Country     string    `gorm:"type:varchar(100);default:'Indonesia'" json:"country"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *ClientProfile) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	DisplayName string       `gorm:"type:varchar(100);not null" json:"display_name"`
	Description *string      `gorm:"type:text" json:"description,omitempty"`
	IsSystem    bool         `gorm:"not null" json:"is_system"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

func (r *Role) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time `json:"created_at"`
}

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"type:varchar(150);not null" json:"display_name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (p *Permission) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       *uuid.UUID `gorm:"type:uuid;index" json:"user_id,omitempty"`
	Action       string     `gorm:"type:varchar(100);not null" json:"action"`
	ResourceType string     `gorm:"type:varchar(100);not null" json:"resource_type"`
	ResourceID   *uuid.UUID `gorm:"type:uuid" json:"resource_id,omitempty"`
	OldValues    *string    `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues    *string    `gorm:"type:jsonb" json:"new_values,omitempty"`
	IPAddress    *string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent    *string    `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
