package http

import (
	"backend-of/internal/auth"
	"backend-of/internal/config"
	"backend-of/internal/domain"
	"backend-of/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg     config.Config
	authSvc *service.AuthService
	simSvc  *service.SimulationService
	limiter *auth.LoginLimiter
}

func NewHandler(cfg config.Config, authSvc *service.AuthService, simSvc *service.SimulationService, limiter *auth.LoginLimiter) *Handler {
	return &Handler{cfg: cfg, authSvc: authSvc, simSvc: simSvc, limiter: limiter}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(c *gin.Context) {
	if !h.limiter.Allow(c.ClientIP()) {
		c.JSON(http.StatusTooManyRequests, domain.ErrorResponse{Error: domain.NewError("rate_limited", "demasiados intentos de login", "")})
		return
	}
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "username y password son requeridos", "")})
		return
	}
	acc, ref, err := h.authSvc.Login(c, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: domain.NewError("unauthorized", "credenciales invalidas", "")})
		return
	}
	h.setAuthCookies(c, acc, ref)
	c.JSON(http.StatusOK, gin.H{"user": gin.H{"username": req.Username}})
}

func (h *Handler) Logout(c *gin.Context) {
	h.clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Refresh(c *gin.Context) {
	ref, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: domain.NewError("unauthorized", "refresh token no encontrado", "")})
		return
	}
	acc, newRef, err := h.authSvc.Refresh(ref)
	if err != nil {
		c.JSON(http.StatusUnauthorized, domain.ErrorResponse{Error: domain.NewError("unauthorized", "refresh token invalido", "")})
		return
	}
	h.setAuthCookies(c, acc, newRef)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) Session(c *gin.Context) {
	u, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{"authenticated": true, "user": gin.H{"username": u}})
}

func (h *Handler) CreateSimulation(c *gin.Context) {
	var in domain.SimulacionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "payload invalido", "")})
		return
	}
	if errs := service.ValidateSimulationInput(in); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "errores de validacion", ""), Errors: errs})
		return
	}
	sim, err := h.simSvc.Create(c, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error creando simulacion", "")})
		return
	}
	c.JSON(http.StatusCreated, sim)
}

func (h *Handler) ListSimulations(c *gin.Context) {
	nombre := c.Query("cliente")
	items, err := h.simSvc.ListByNombreCliente(c, nombre)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error listando simulaciones", "")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetSimulationByID(c *gin.Context) {
	id := c.Param("id")
	sim, ok, err := h.simSvc.GetByID(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error consultando simulacion", "")})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: domain.NewError("not_found", "simulacion no encontrada", "id")})
		return
	}
	c.JSON(http.StatusOK, sim)
}

func (h *Handler) setAuthCookies(c *gin.Context, access, refresh string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("access_token", access, int(h.cfg.AccessTTL.Seconds()), "/", "", h.cfg.CookieSecure, true)
	c.SetCookie("refresh_token", refresh, int(h.cfg.RefreshTTL.Seconds()), "/", "", h.cfg.CookieSecure, true)
}

func (h *Handler) clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", h.cfg.CookieSecure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", h.cfg.CookieSecure, true)
}
