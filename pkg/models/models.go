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
	ClientProfile   *ClientProfile `gorm:"foreignKey:UserID" json:"client_profile,omitempty"`
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
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	DisplayName string         `gorm:"type:varchar(100);not null" json:"display_name"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	IsSystem    bool           `gorm:"not null" json:"is_system"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
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

type CompanySetting struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name          string    `gorm:"type:varchar(255);not null" json:"name"`
	Address       string    `gorm:"type:text" json:"address"`
	Description   string    `gorm:"type:text" json:"description"`
	FaviconURL    string    `gorm:"type:varchar(255)" json:"favicon_url"`
	LogoLongURL   string    `gorm:"type:varchar(255)" json:"logo_long_url"`
	LogoSquareURL string    `gorm:"type:varchar(255)" json:"logo_square_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (c *CompanySetting) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type FeaturesConfig map[string]interface{}

type Package struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name           string         `gorm:"type:varchar(255);not null" json:"name"`
	Price          float64        `gorm:"type:numeric(15,2);not null" json:"price"`
	FeaturesConfig FeaturesConfig `gorm:"type:jsonb;serializer:json" json:"features_config"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (p *Package) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

type Template struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	NuxtComponent string         `gorm:"type:varchar(255);not null" json:"nuxt_component"`
	ThumbnailURL  *string        `gorm:"type:varchar(255)" json:"thumbnail_url,omitempty"`
	IsActive      bool           `gorm:"not null;default:true" json:"is_active"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (t *Template) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type Feature struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	FeatureKey   string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"feature_key"`
	FeatureName  string         `gorm:"type:varchar(255);not null" json:"feature_name"`
	InputType    string         `gorm:"type:varchar(50);not null" json:"input_type"`
	DefaultValue string         `gorm:"type:varchar(255);not null" json:"default_value"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (f *Feature) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

type OrderStatus string

const (
	OrderStatusUnpaid  OrderStatus = "unpaid"
	OrderStatusPaid    OrderStatus = "paid"
	OrderStatusExpired OrderStatus = "expired"
)

type Client struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Email     string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Whatsapp  string         `gorm:"type:varchar(50);not null" json:"whatsapp"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Orders    []Order        `gorm:"foreignKey:ClientID" json:"orders"`
}

func (c *Client) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

type Order struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	InvoiceNumber string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"invoice_number"`
	ClientID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"client_id"`
	PackageID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"package_id"`
	TotalAmount   float64        `gorm:"type:numeric(15,2);not null" json:"total_amount"`
	Status        OrderStatus    `gorm:"type:varchar(20);not null;default:'unpaid'" json:"status"`
	PaymentURL    string         `gorm:"type:text" json:"payment_url"`
	FormToken     *string        `gorm:"type:varchar(255);index" json:"form_token"`
	ScannerToken  *string        `gorm:"type:varchar(255);index" json:"scanner_token"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	Client  *Client  `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Package *Package `gorm:"foreignKey:PackageID" json:"package,omitempty"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

type Invitation struct {
	ID          uuid.UUID              `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID     uuid.UUID              `gorm:"type:uuid;not null;uniqueIndex" json:"order_id"`
	ClientID    uuid.UUID              `gorm:"type:uuid;not null;index" json:"client_id"`
	PackageID   uuid.UUID              `gorm:"type:uuid;not null;index" json:"package_id"`
	Title       string                 `gorm:"type:varchar(255);not null;default:'Draft Undangan'" json:"title"`
	Slug        *string                `gorm:"type:varchar(255);uniqueIndex" json:"slug,omitempty"`
	Status      string                 `gorm:"type:varchar(50);not null;default:'draft'" json:"status"`
	GroomData   map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"groom_data,omitempty"`
	BrideData   map[string]interface{} `gorm:"type:jsonb;serializer:json" json:"bride_data,omitempty"`
	EventsData  interface{}            `gorm:"type:jsonb;serializer:json" json:"events_data,omitempty"`
	StoryData   interface{}            `gorm:"type:jsonb;serializer:json" json:"story_data,omitempty"`
	GalleryURLs []string               `gorm:"type:jsonb;serializer:json" json:"gallery_urls,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   gorm.DeletedAt         `gorm:"index" json:"-"`

	Order   *Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Client  *Client  `gorm:"foreignKey:ClientID" json:"client,omitempty"`
	Package *Package `gorm:"foreignKey:PackageID" json:"package,omitempty"`
	Guests  []Guest  `gorm:"foreignKey:InvitationID" json:"guests,omitempty"`
}

func (i *Invitation) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

type Guest struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	InvitationID     uuid.UUID      `gorm:"type:uuid;not null;index" json:"invitation_id"`
	Name             string         `gorm:"type:varchar(255);not null" json:"name"`
	Phone            *string        `gorm:"type:varchar(50)" json:"phone,omitempty"`
	Pax              int            `gorm:"default:1" json:"pax"`
	QRToken          string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"qr_token"`
	RSVPStatus       string         `gorm:"type:varchar(20);default:'pending'" json:"rsvp_status"` // 'hadir', 'tidak_hadir', 'pending'
	ActualAttendance bool           `gorm:"default:false" json:"actual_attendance"`
	AttendanceTime   *time.Time     `gorm:"type:timestamptz" json:"attendance_time,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	Invitation *Invitation `gorm:"foreignKey:InvitationID" json:"invitation,omitempty"`
}

func (g *Guest) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

