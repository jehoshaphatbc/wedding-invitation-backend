package handler

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/config"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/database"
	"github.com/jehoshaphatbc/wedding-invitation-backend/pkg/server"
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
	// Add CORS headers for Vercel edge cases when init fails
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Seeder-Secret")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

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
