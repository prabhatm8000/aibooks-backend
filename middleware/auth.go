package middleware

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func IsAuthenticated(c *gin.Context) {
	tokenString, err := c.Cookie("auth-token")
	if err != nil {
		c.IndentedJSON(401, gin.H{"message": "Failed to retrieve token."})
		c.Abort()
		return
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		c.IndentedJSON(401, gin.H{"message": "Failed to retrieve token."})
		c.Abort()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["user_id"].(string)
		c.Set("user_id", userId)
	} else {
		c.IndentedJSON(401, gin.H{"message": "Invalid token"})
		c.Abort()
		return
	}

	c.Next()
}
