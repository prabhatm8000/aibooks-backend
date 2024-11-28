package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	apiRoutes := r.Group("/api/v1")

	RegisterUserRoutes(apiRoutes)
	RegisterAuthRoutes(apiRoutes)
	RegisterBookdataRoutes(apiRoutes)
	RegisterStaticDataRoutes(apiRoutes)
}
