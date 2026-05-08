package http

import (
	"backend-of/internal/auth"
	"backend-of/internal/config"
	"backend-of/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, authSvc *service.AuthService, simSvc *service.SimulationService, vehicleSvc *service.VehicleService) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware(cfg.FrontendOrigins))
	h := NewHandler(cfg, authSvc, simSvc, vehicleSvc, auth.NewLoginLimiter(cfg.LoginMaxPerMinute))

	api := r.Group("/api/v1")
	authGroup := api.Group("/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
	authGroup.POST("/logout", h.Logout)
	authGroup.POST("/refresh", h.Refresh)
	authGroup.GET("/session", RequireAuth(authSvc), h.Session)
	api.GET("/bancos", h.ListBanks)

	simGroup := api.Group("/simulaciones", RequireAuth(authSvc))
	simGroup.POST("", h.CreateSimulation)
	simGroup.GET("", h.ListSimulations)
	simGroup.GET("/:id", h.GetSimulationByID)

	vehicleGroup := api.Group("/vehiculos", RequireAuth(authSvc))
	vehicleGroup.POST("", h.CreateVehicle)
	vehicleGroup.GET("", h.ListVehicles)
	vehicleGroup.GET("/:id", h.GetVehicleByID)

	clientGroup := api.Group("/clientes", RequireAuth(authSvc))
	clientGroup.GET("/me", h.GetClientProfile)

	r.GET("/openapi.json", h.OpenAPI)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func corsMiddleware(originsCSV string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, origin := range strings.Split(originsCSV, ",") {
		o := strings.TrimSpace(origin)
		if o != "" {
			allowed[o] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if _, ok := allowed[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
