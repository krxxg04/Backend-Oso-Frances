package main

import (
	"backend-of/internal/application/common/config"
	"backend-of/internal/application/services"
	"backend-of/internal/infrastructure/db/jsondb"
	"backend-of/internal/infrastructure/db/postgres"
	"backend-of/internal/infrastructure/security"
	restapi "backend-of/internal/interface/api/rest"
	"context"
	"log"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	tokens := security.NewTokenManager(cfg.JWTSecret)

	var authSvc *services.AuthService
	var simSvc *services.SimulationService
	var vehicleSvc *services.VehicleService

	if cfg.DatabaseURL != "" {
		log.Println("storage mode: PostgreSQL (DATABASE_URL)")
		store, err := postgres.NewStore(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("error connecting postgres: %v", err)
		}
		defer store.Close()
		authSvc = services.NewAuthService(store, store, tokens, cfg.AccessTTL, cfg.RefreshTTL)
		simSvc = services.NewSimulationService(store, store)
		vehicleSvc = services.NewVehicleService(store, store)
	} else {
		log.Printf("storage mode: JSON local (%s)", cfg.DataFilePath)
		store, err := jsondb.NewStore(cfg.DataFilePath)
		if err != nil {
			log.Fatalf("error opening data store: %v", err)
		}
		authSvc = services.NewAuthService(store, store, tokens, cfg.AccessTTL, cfg.RefreshTTL)
		simSvc = services.NewSimulationService(store, store)
		vehicleSvc = services.NewVehicleService(store, store)
	}

	if err := authSvc.Seed(ctx); err != nil {
		log.Fatalf("error seeding users: %v", err)
	}

	r := restapi.NewRouter(cfg, authSvc, simSvc, vehicleSvc)
	log.Printf("server starting on :%s", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
