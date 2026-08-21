//go:build ignore

package main

import (
	"log"
	"net/http"

	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/database"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/server"
	seeds "github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

var Handler http.Handler

func init() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db := database.Connect(cfg)
	seeds.Seed(db, cfg.SuperAdminName, cfg.SuperAdminEmail, cfg.SuperAdminPassword)

	Handler = server.New(cfg, db)
}

func main() {}
