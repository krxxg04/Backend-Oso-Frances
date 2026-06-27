package rest

import (
	"backend-of/internal/application/common/config"
	"backend-of/internal/application/services"
	domain "backend-of/internal/domain/entities"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(cfg config.Config, auth *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err == nil {
			u, verifyErr := auth.VerifyAccess(token)
			if verifyErr == nil {
				c.Set("username", u)
				c.Next()
				return
			}
		}

		refreshToken, refreshErr := c.Cookie("refresh_token")
		if refreshErr == nil {
			username, access, refresh, rotateErr := auth.RefreshWithSubject(refreshToken)
			if rotateErr == nil {
				c.SetSameSite(cookieSameSiteMode(cfg.CookieSameSite))
				c.SetCookie("access_token", access, int(cfg.AccessTTL.Seconds()), "/", "", cfg.CookieSecure, true)
				c.SetCookie("refresh_token", refresh, int(cfg.RefreshTTL.Seconds()), "/", "", cfg.CookieSecure, true)
				c.Set("username", username)
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: domain.NewError("unauthorized", "sesion no encontrada", "")})
	}
}
