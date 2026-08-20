package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/models"
)

func Connect(cfg *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if cfg.IsProduction() {
		fmt.Println("Production mode: skipping DROP TABLE and AutoMigrate (use migrate CLI)")
	} else {
		db.Exec("DROP TABLE IF EXISTS audit_logs CASCADE")
		db.Exec("DROP TABLE IF EXISTS email_verification_tokens CASCADE")
		db.Exec("DROP TABLE IF EXISTS password_reset_tokens CASCADE")
		db.Exec("DROP TABLE IF EXISTS refresh_tokens CASCADE")
		db.Exec("DROP TABLE IF EXISTS role_permissions CASCADE")
		db.Exec("DROP TABLE IF EXISTS user_roles CASCADE")
		db.Exec("DROP TABLE IF EXISTS client_profiles CASCADE")
		db.Exec("DROP TABLE IF EXISTS permissions CASCADE")
		db.Exec("DROP TABLE IF EXISTS roles CASCADE")
		db.Exec("DROP TABLE IF EXISTS users CASCADE")

		fmt.Println("Old tables dropped")

		err = db.AutoMigrate(
			&models.User{},
			&models.ClientProfile{},
			&models.Role{},
			&models.Permission{},
			&models.UserRole{},
			&models.RolePermission{},
			&models.RefreshToken{},
			&models.PasswordResetToken{},
			&models.EmailVerificationToken{},
			&models.AuditLog{},
		)
		if err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}

		fmt.Println("Database connected and migrated successfully")
	}

	return db
}
