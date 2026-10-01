package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/models"
	"github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.GetDSN()
	if dsn == "" {
		return nil, fmt.Errorf("database DSN is empty")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.ClientProfile{},
		&models.Role{},
		&models.Permission{},
		&models.CompanySetting{},
		&models.UserRole{},
		&models.RolePermission{},
		&models.RefreshToken{},
		&models.PasswordResetToken{},
		&models.EmailVerificationToken{},
		&models.AuditLog{},
		&models.Package{},
		&models.Template{},
		&models.Feature{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Auto-seed default features if table is empty
	seeds.SeedFeatures(db)

	fmt.Println("Database connected and migrated successfully")

	return db, nil
}
