package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyRole   = "user_role"
)

// AuthRequired validates Bearer JWT in Authorization header.
func AuthRequired() gin.HandlerFunc {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev_secret_please_change"
	}
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" || !strings.HasPrefix(h, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		raw := strings.TrimPrefix(h, "Bearer ")
		tok, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !tok.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		claims := tok.Claims.(jwt.MapClaims)
		c.Set(ContextKeyUserID, claims["sub"].(string))
		c.Set(ContextKeyRole, claims["role"].(string))
		c.Next()
	}
}

// OrganizerRequired ensures role == "organizer".
func OrganizerRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(ContextKeyRole)
		if role.(string) != "organizer" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "organizer role required"})
			return
		}
		c.Next()
	}
}
