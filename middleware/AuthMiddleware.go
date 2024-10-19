package middleware

import (
	"nerde_yenir/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		ClientToken := c.Request.Header.Get("Authorization")

		if ClientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token doğrulanamadı"})
			c.Abort()
			return
		}
		claims, err := jwt.ValidateToken(ClientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		c.Set("uid", claims.UserId)
		c.Next()
	}
}
