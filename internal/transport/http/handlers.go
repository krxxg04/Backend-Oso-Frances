package http

import (
	"backend-of/internal/auth"
	"backend-of/internal/config"
	"backend-of/internal/domain"
	"backend-of/internal/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg     config.Config
	authSvc *service.AuthService
	simSvc  *service.SimulationService
	vehSvc  *service.VehicleService
	limiter *auth.LoginLimiter
}

func NewHandler(cfg config.Config, authSvc *service.AuthService, simSvc *service.SimulationService, vehSvc *service.VehicleService, limiter *auth.LoginLimiter) *Handler {
	return &Handler{cfg: cfg, authSvc: authSvc, simSvc: simSvc, vehSvc: vehSvc, limiter: limiter}
}

type loginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type registerReq struct {
	Username       string `json:"username" binding:"required"`
	Gmail          string `json:"gmail" binding:"required"`
	DNI            string `json:"dni" binding:"required"`
	Password       string `json:"password" binding:"required"`
	RepeatPassword string `json:"repeatPassword" binding:"required"`
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

func (h *Handler) Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "username, gmail, dni, password y repeatPassword son requeridos", "")})
		return
	}
	if req.Password != req.RepeatPassword {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "password y repeatPassword deben coincidir", "repeatPassword")})
		return
	}
	acc, ref, err := h.authSvc.RegisterProfile(c, domain.User{
		Username: req.Username,
		Email:    req.Gmail,
		DNI:      req.DNI,
	}, req.Password)
	if err != nil {
		switch err.Error() {
		case "validation_error":
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "username minimo 3 chars, sin espacios; password minimo 6 chars; DNI 8 digitos; email valido", "")})
		case "conflict":
			c.JSON(http.StatusConflict, domain.ErrorResponse{Error: domain.NewError("conflict", "el username ya existe", "username")})
		default:
			c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error registrando usuario", "")})
		}
		return
	}
	h.setAuthCookies(c, acc, ref)
	c.JSON(http.StatusCreated, gin.H{"user": gin.H{"username": req.Username, "gmail": req.Gmail, "dni": req.DNI}})
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
	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)
	in.NombreCliente = username
	in, errs := service.ApplySimulationRules(in)
	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "errores de validacion", ""), Errors: errs})
		return
	}
	sim, err := h.simSvc.CreateForUser(c, username, in)
	if err != nil {
		if err.Error() == "validation_error" {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "errores de validacion", "")})
			return
		}
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error creando simulacion", "")})
		return
	}
	c.JSON(http.StatusCreated, sim)
}

func (h *Handler) ListSimulations(c *gin.Context) {
	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)
	items, err := h.simSvc.ListByUserFiltered(c, username, simulationFilterFromQuery(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error listando simulaciones", "")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetSimulationByID(c *gin.Context) {
	id := c.Param("id")
	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)
	sim, ok, err := h.simSvc.GetByIDForUser(c, username, id)
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

func (h *Handler) CreateVehicle(c *gin.Context) {
	var in domain.Vehicle
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "payload invalido", "")})
		return
	}
	if errs := service.ValidateVehicle(in); len(errs) > 0 {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: domain.NewError("validation_error", "errores de validacion", ""), Errors: errs})
		return
	}
	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)
	vehicle, err := h.vehSvc.CreateForUser(c, username, in)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error creando vehiculo", "")})
		return
	}
	c.JSON(http.StatusCreated, vehicle)
}

func (h *Handler) ListVehicles(c *gin.Context) {
	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)
	items, err := h.vehSvc.ListByUser(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error listando vehiculos", "")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetVehicleByID(c *gin.Context) {
	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)
	vehicle, ok, err := h.vehSvc.GetByIDForUser(c, username, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error consultando vehiculo", "")})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: domain.NewError("not_found", "vehiculo no encontrado", "id")})
		return
	}
	c.JSON(http.StatusOK, vehicle)
}

func (h *Handler) GetClientProfile(c *gin.Context) {
	usernameAny, _ := c.Get("username")
	username, _ := usernameAny.(string)
	user, ok, err := h.authSvc.GetUser(c, username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: domain.NewError("internal_error", "error consultando cliente", "")})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: domain.NewError("not_found", "cliente no encontrado", "username")})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": user.Username, "email": user.Email, "dni": user.DNI, "fullName": user.FullName, "role": user.Role})
}

func (h *Handler) OpenAPI(c *gin.Context) {
	c.JSON(http.StatusOK, openAPISpec())
}

func (h *Handler) ListBankOptions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": service.BankOptions()})
}

func (h *Handler) setAuthCookies(c *gin.Context, access, refresh string) {
	c.SetSameSite(cookieSameSiteMode(h.cfg.CookieSameSite))
	c.SetCookie("access_token", access, int(h.cfg.AccessTTL.Seconds()), "/", "", h.cfg.CookieSecure, true)
	c.SetCookie("refresh_token", refresh, int(h.cfg.RefreshTTL.Seconds()), "/", "", h.cfg.CookieSecure, true)
}

func (h *Handler) clearAuthCookies(c *gin.Context) {
	c.SetSameSite(cookieSameSiteMode(h.cfg.CookieSameSite))
	c.SetCookie("access_token", "", -1, "/", "", h.cfg.CookieSecure, true)
	c.SetCookie("refresh_token", "", -1, "/", "", h.cfg.CookieSecure, true)
}

func cookieSameSiteMode(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

func simulationFilterFromQuery(c *gin.Context) domain.SimulacionFilter {
	return domain.SimulacionFilter{
		FechaDesde: c.Query("fechaDesde"),
		FechaHasta: c.Query("fechaHasta"),
		Moneda:     domain.Currency(c.Query("moneda")),
		PlazoMeses: queryInt(c, "plazoMeses"),
		MontoMin:   queryFloat(c, "montoMin"),
		MontoMax:   queryFloat(c, "montoMax"),
		Vehiculo:   c.Query("vehiculo"),
	}
}

func queryInt(c *gin.Context, key string) int {
	raw := c.Query(key)
	if raw == "" {
		return 0
	}
	n, _ := strconv.Atoi(raw)
	return n
}

func queryFloat(c *gin.Context, key string) float64 {
	raw := c.Query(key)
	if raw == "" {
		return 0
	}
	n, _ := strconv.ParseFloat(raw, 64)
	return n
}
