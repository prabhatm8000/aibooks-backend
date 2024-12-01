package middleware

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func IsAuthenticated(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("user_id")

	if userId == nil {
		c.IndentedJSON(401, gin.H{"message": "Sign in first."})
		c.Abort()
		return
	}

	c.Next()
}
