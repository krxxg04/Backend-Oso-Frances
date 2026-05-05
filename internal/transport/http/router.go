package http

import (
	"backend-of/internal/auth"
	"backend-of/internal/config"
	"backend-of/internal/service"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, authSvc *service.AuthService, simSvc *service.SimulationService) *gin.Engine {
	r := gin.Default()
	h := NewHandler(cfg, authSvc, simSvc, auth.NewLoginLimiter(cfg.LoginMaxPerMinute))

	api := r.Group("/api/v1")
	authGroup := api.Group("/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
	authGroup.POST("/logout", h.Logout)
	authGroup.POST("/refresh", h.Refresh)
	authGroup.GET("/session", RequireAuth(authSvc), h.Session)

	simGroup := api.Group("/simulaciones", RequireAuth(authSvc))
	simGroup.POST("", h.CreateSimulation)
	simGroup.GET("", h.ListSimulations)
	simGroup.GET("/:id", h.GetSimulationByID)

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}
