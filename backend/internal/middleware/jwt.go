package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"suxin/internal/appctx"
	"suxin/internal/pkg/jwtx"
)

func AuthRequired(ctx *appctx.AppContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}
		token := strings.TrimSpace(auth[len("Bearer "):])
		claims, err := jwtx.Parse(token, ctx.Config.Auth.JWTSecret)
		if err != nil || claims.Typ != "access" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}
