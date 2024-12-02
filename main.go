package main

import (
	"example/aibooks-backend/config"
	"example/aibooks-backend/routes"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if os.Getenv("ENV") != "PROD" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalln(".env file not found.")
		}
	} else {
		log.Println("Running in PROD mode.")
	}

	disconnectMongoDB := config.ConnectMongoDB()
	defer disconnectMongoDB()

	ginMode := os.Getenv("GIN_MODE")
	gin.SetMode(ginMode)
	var frontend string
	switch ginMode {
	case "release":
		frontend = os.Getenv("FRONTEND_PROD_URL")
	default:
		frontend = os.Getenv("FRONTEND_DEV_URL")
	}
	router := gin.Default()

	corsConfigs := cors.Config{
		AllowOrigins:     []string{frontend},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Accept", "Origin", "X-Requested-With"},
		AllowCredentials: true, // Only works with specific origins, not "*"
	}
	router.Use(cors.New(corsConfigs))

	log.Println("Allowed origin:", frontend)

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	routes.RegisterRoutes(router)

	// test route
	router.GET("/", func(c *gin.Context) {
		c.IndentedJSON(200, gin.H{
			"message": "API works fine!",
		})
	})

	router.Run("0.0.0.0:8080")
}
