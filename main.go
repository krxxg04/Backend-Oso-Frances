package main

import (
	"backend-of/internal/auth"
	"backend-of/internal/config"
	"backend-of/internal/repository/jsondb"
	"backend-of/internal/repository/postgres"
	"backend-of/internal/service"
	httptransport "backend-of/internal/transport/http"
	"context"
	"log"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	tokens := auth.NewTokenManager(cfg.JWTSecret)

	var authSvc *service.AuthService
	var simSvc *service.SimulationService

	if cfg.DatabaseURL != "" {
		log.Println("storage mode: PostgreSQL (DATABASE_URL)")
		store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("error connecting postgres: %v", err)
		}
		defer store.Close()
		authSvc = service.NewAuthService(store, store, tokens, cfg.AccessTTL, cfg.RefreshTTL)
		simSvc = service.NewSimulationService(store, store)
	} else {
		log.Printf("storage mode: JSON local (%s)", cfg.DataFilePath)
		store, err := jsondb.NewStore(cfg.DataFilePath)
		if err != nil {
			log.Fatalf("error opening data store: %v", err)
		}
		authSvc = service.NewAuthService(store, store, tokens, cfg.AccessTTL, cfg.RefreshTTL)
		simSvc = service.NewSimulationService(store, store)
	}

	if err := authSvc.Seed(ctx); err != nil {
		log.Fatalf("error seeding users: %v", err)
	}

	r := httptransport.NewRouter(cfg, authSvc, simSvc)
	log.Printf("server starting on :%s", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
