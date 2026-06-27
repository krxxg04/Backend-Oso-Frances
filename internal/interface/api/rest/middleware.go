package rest

import (
	"backend-of/internal/application/services"
	domain "backend-of/internal/domain/entities"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireAuth(auth *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: domain.NewError("unauthorized", "sesion no encontrada", "")})
			return
		}
		u, err := auth.VerifyAccess(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, domain.ErrorResponse{Error: domain.NewError("unauthorized", "sesion invalida", "")})
			return
		}
		c.Set("username", u)
		c.Next()
	}
}
