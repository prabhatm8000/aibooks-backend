package main

import (
	"example/aibooks-backend/config"
	"example/aibooks-backend/routes"
	"log"
	"net/http"
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
	frontendProd := os.Getenv("FRONTEND_PROD_URL")
	frontendDev := os.Getenv("FRONTEND_DEV_URL")
	router := gin.Default()

	corsConfigs := cors.Config{
		AllowOrigins:     []string{frontendProd, frontendDev},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Accept", "Origin", "X-Requested-With"},
		AllowCredentials: true, // Only works with specific origins, not "*"
	}
	router.Use(cors.New(corsConfigs))

	log.Printf("Allowed origin: %s, %s\n", frontendProd, frontendDev)

	secretKey := os.Getenv("SESSION_SECRET")
	if secretKey == "" {
		log.Fatalln("SESSION_SECRET not set")
	}
	store := cookie.NewStore([]byte(secretKey))
	store.Options(sessions.Options{
		Path:     "/",
		HttpOnly: true,
		Secure:   ginMode == "release",
		SameSite: http.SameSiteStrictMode,
	})
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
