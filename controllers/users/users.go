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
		c.Error(err)
		return
	}

	c.IndentedJSON(200, user)
}
