package users

import (
	"example/aibooks-backend/models/users"

	"github.com/gin-gonic/gin"
)

func GetUser(c *gin.Context) {
	userId := c.GetString("user_id")

	user, err := users.GetUserById(userId)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": "Uh oh! Something went wrong."})
		return
	}

	c.IndentedJSON(200, user)
}
