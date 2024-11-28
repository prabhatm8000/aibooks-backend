package staticdatas

import (
	"example/aibooks-backend/models/staticdatas"

	"github.com/gin-gonic/gin"
)

func GetStaticDataByDataType(c *gin.Context) {
	dataType := c.Param("dataType")

	staticData, err := staticdatas.GetStaticDataByType(dataType)
	if err != nil {
		c.IndentedJSON(400, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(200, staticData)
}
