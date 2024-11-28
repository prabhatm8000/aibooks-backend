package routes

import (
	"example/aibooks-backend/controllers/staticdatas"

	"github.com/gin-gonic/gin"
)

func RegisterStaticDataRoutes(r *gin.RouterGroup) {
	staticDataGroup := r.Group("/staticData")

	{
		staticDataGroup.GET("/:dataType", staticdatas.GetStaticDataByDataType)
	}
}
