package main

import (
	"backend-of/internal/auth"
	"backend-of/internal/config"
	"backend-of/internal/repository/jsondb"
	"backend-of/internal/service"
	httptransport "backend-of/internal/transport/http"
	"log"
)

func main() {
	cfg := config.Load()

	store, err := jsondb.NewStore(cfg.DataFilePath)
	if err != nil {
		log.Fatalf("error opening data store: %v", err)
	}

	authSvc := service.NewAuthService(store, auth.NewTokenManager(cfg.JWTSecret), cfg.AccessTTL, cfg.RefreshTTL)
	if err := authSvc.Seed(nil); err != nil {
		log.Fatalf("error seeding users: %v", err)
	}

	simSvc := service.NewSimulationService(store, store)
	r := httptransport.NewRouter(cfg, authSvc, simSvc)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
