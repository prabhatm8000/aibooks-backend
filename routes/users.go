package routes

import (
	"example/aibooks-backend/controllers/users"
	"example/aibooks-backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.RouterGroup) {
	usersGroup := r.Group("/users")
	usersGroup.Use(middleware.IsAuthenticated)
	{
		usersGroup.GET("/", users.GetUser)
	}
}
