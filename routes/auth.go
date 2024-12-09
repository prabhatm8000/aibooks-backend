package routes

import (
	"example/aibooks-backend/controllers/auth"
	"example/aibooks-backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup) {
	authGrp := r.Group("/auth")

	{
		authGrp.POST("/login", auth.Login)

		authGrp.POST("/sendOtp", auth.SendOtp)

		authGrp.POST("/create", auth.CreateAccount)

		authGrp.Use(middleware.IsAuthenticated).GET("/logout", auth.Logout)

		authGrp.Use(middleware.IsAuthenticated).GET("/user", auth.GetUserDetails)
	}
}
