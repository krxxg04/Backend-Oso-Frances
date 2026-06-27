package rest

import (
	"backend-of/internal/application/common/config"
	"backend-of/internal/application/services"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func NewRouter(cfg config.Config, authSvc *services.AuthService, simSvc *services.SimulationService, vehicleSvc *services.VehicleService) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware(cfg.FrontendOrigins))
	h := NewHandler(cfg, authSvc, simSvc, vehicleSvc, NewLoginLimiter(cfg.LoginMaxPerMinute))
	r.GET("/", func(c *gin.Context) { c.Redirect(http.StatusFound, "/swagger") })

	api := r.Group("/api/v1")
	api.GET("/bancos", h.ListBankOptions)

	authGroup := api.Group("/auth")
	authGroup.POST("/register", h.Register)
	authGroup.POST("/login", h.Login)
	authGroup.POST("/logout", h.Logout)
	authGroup.POST("/refresh", h.Refresh)
	authGroup.GET("/google/login", h.GoogleLogin)
	authGroup.GET("/google/callback", h.GoogleCallback)
	authGroup.GET("/session", RequireAuth(authSvc), h.Session)

	simGroup := api.Group("/simulaciones", RequireAuth(authSvc))
	simGroup.POST("", h.CreateSimulation)
	simGroup.GET("", h.ListSimulations)
	simGroup.GET("/:id", h.GetSimulationByID)

	vehicleGroup := api.Group("/vehiculos", RequireAuth(authSvc))
	vehicleGroup.POST("", h.CreateVehicle)
	vehicleGroup.GET("", h.ListVehicles)
	vehicleGroup.GET("/:id", h.GetVehicleByID)
	vehicleGroup.PUT("/:id", h.UpdateVehicle)

	clientGroup := api.Group("/clientes", RequireAuth(authSvc))
	clientGroup.GET("/me", h.GetClientProfile)
	clientGroup.PUT("/me", h.UpdateClientProfile)

	r.GET("/openapi.json", h.OpenAPI)
	r.GET("/swagger", swaggerUI)
	r.GET("/swagger/", swaggerUI)
	r.GET("/docs", swaggerUI)
	r.GET("/docs/", swaggerUI)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	return r
}

func corsMiddleware(originsCSV string) gin.HandlerFunc {
	allowed := make([]string, 0)
	for _, origin := range strings.Split(originsCSV, ",") {
		o := normalizeOrigin(strings.TrimSpace(origin))
		if o != "" {
			allowed = append(allowed, o)
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" {
			if isAllowedOrigin(origin, allowed) {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin, Access-Control-Request-Method, Access-Control-Request-Headers")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Headers", allowedHeaders(c.GetHeader("Access-Control-Request-Headers")))
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
			}
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	normalizedOrigin := normalizeOrigin(origin)
	u, err := url.Parse(normalizedOrigin)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, a := range allowed {
		if normalizedOrigin == a {
			return true
		}
		parsedAllowed, err := url.Parse(a)
		if err != nil || parsedAllowed.Hostname() == "" || parsedAllowed.Scheme == "" {
			continue
		}
		allowedHost := strings.ToLower(parsedAllowed.Hostname())
		if strings.HasPrefix(allowedHost, "*.") &&
			u.Scheme == parsedAllowed.Scheme &&
			strings.HasSuffix(host, strings.TrimPrefix(allowedHost, "*.")) {
			return true
		}
	}
	return false
}

func normalizeOrigin(origin string) string {
	return strings.TrimRight(strings.TrimSpace(origin), "/")
}

func allowedHeaders(requested string) string {
	if strings.TrimSpace(requested) == "" {
		return "Accept, Authorization, Content-Type, X-Requested-With"
	}
	return requested
}
