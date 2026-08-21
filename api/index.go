package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/database"
	"github.com/jehoshaphatbc/wedding-invitation-backend/internal/server"
	seeds "github.com/jehoshaphatbc/wedding-invitation-backend/seeds"
)

var (
	engine  *gin.Engine
	once    sync.Once
	initErr error
)

func initialize() {
	once.Do(func() {
		cfg, err := config.Load()
		if err != nil {
			initErr = err
			log.Printf("Failed to load config: %v", err)
			return
		}

		db, err := database.Connect(cfg)
		if err != nil {
			initErr = err
			log.Printf("Failed to connect database: %v", err)
			return
		}

		seeds.Seed(db, cfg.SuperAdminName, cfg.SuperAdminEmail, cfg.SuperAdminPassword)

		engine = server.New(cfg, db)
	})
}

func Handler(w http.ResponseWriter, r *http.Request) {
	initialize()

	if initErr != nil {
		http.Error(w, "Service unavailable: "+initErr.Error(), http.StatusServiceUnavailable)
		return
	}

	if engine == nil {
		http.Error(w, "Server not initialized", http.StatusServiceUnavailable)
		return
	}

	engine.ServeHTTP(w, r)
}
