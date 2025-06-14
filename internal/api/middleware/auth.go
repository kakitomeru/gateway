package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/kakitomeru/shared/jwt"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(protectedPrefixes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		isProtected := false
		path := c.Request.URL.Path

		for _, prefix := range protectedPrefixes {
			if strings.HasPrefix(path, prefix) {
				isProtected = true
				break
			}
		}

		if !isProtected {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no authorization header provided"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		token := parts[1]
		ctx := context.WithValue(c.Request.Context(), "accessToken", token)

		claims, err := jwt.ValidateJWT(token)
		if err != nil {
			if errors.Is(err, jwt.ErrInvalidToken) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			} else if errors.Is(err, jwt.ErrExpiredToken) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "expired token"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "something went wrong, try again later"})
			}
			c.Abort()
			return
		}

		ctx = context.WithValue(ctx, "userID", claims.UserID.String())
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
