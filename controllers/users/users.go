package users

import (
	"example/aibooks-backend/models/users"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	session := sessions.Default(c)
	userId := session.Get("user_id")

	user, err := users.GetUserById(userId.(string))
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, user)
}
