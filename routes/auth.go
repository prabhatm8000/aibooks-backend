package routes

import (
	"example/aibooks-backend/authenticator"
	"example/aibooks-backend/controllers/auth"
	"log"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(r *gin.RouterGroup) {
	authObj, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to create authenticator: %v", err)
	}

	authGrp := r.Group("/auth")

	{
		authGrp.POST("/login", auth.LoginHandler(authObj))

		authGrp.GET("/callback", auth.CallbackHandler(authObj))

		authGrp.GET("/logout", auth.LogoutHandler)

		authGrp.GET("/user", auth.UserProfileHandler)
	}
}
