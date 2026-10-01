package main

import (
	"log"

	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/database"
	"github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

func main() {
	log.Println("==================================================")
	log.Println("   Harsava Wedding Invitation - Database Seeder   ")
	log.Println("==================================================")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 1. Seed base system (Roles, Permissions, Super Admin)
	seeds.Seed(db, cfg.SuperAdminName, cfg.SuperAdminEmail, cfg.SuperAdminPassword)

	// 2. Seed hierarchical dummy data (Packages, Clients, Orders, Invitations, Guests)
	if _, err := seeds.SeedDummyData(db); err != nil {
		log.Fatalf("Database seeding failed: %v", err)
	}

	log.Println("==================================================")
	log.Println("   All Dummy Data Seeded Successfully!            ")
	log.Println("==================================================")
}
