package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/database"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/server"
	seeds "github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

var engine *gin.Engine

func init() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db := database.Connect(cfg)
	seeds.Seed(db, cfg.SuperAdminName, cfg.SuperAdminEmail, cfg.SuperAdminPassword)

	engine = server.New(cfg, db)
}

func Handler(w http.ResponseWriter, r *http.Request) {
	engine.ServeHTTP(w, r)
}
