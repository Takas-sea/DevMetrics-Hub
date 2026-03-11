package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/Takas-sea/DevMetrics-Hub/db"
	"github.com/Takas-sea/DevMetrics-Hub/handlers"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	conn, err := db.Init()
	if err != nil {
		log.Fatalf("DB init failed: %v\n", err)
	}
	defer conn.Close()

	if err := db.Migrate(conn); err != nil {
		log.Fatalf("DB migrate failed: %v\n", err)
	}

	authHandler := &handlers.AuthHandler{DB: conn}
	activityHandler := handlers.NewActivityHandler()

	router := gin.Default()
	_ = router.SetTrustedProxies([]string{"127.0.0.1"})

	router.Use(corsMiddleware())

	router.GET("/health", healthCheck)

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", authHandler.Logout)
			auth.GET("/callback", authHandler.Callback)
		}

		users := api.Group("/users")
		{
			users.GET("/:id", getUser)
			users.PUT("/:id", updateUser)
		}

		activities := api.Group("/activities")
		{
			activities.GET("/me", activityHandler.GetMyActivities)
			activities.GET("/:userId", getUserActivities)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v\n", err)
	}
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Server is running",
	})
}

func getUser(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"userId":  userID,
		"message": "get user endpoint",
	})
}

func updateUser(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(http.StatusOK, gin.H{
		"userId":  userID,
		"message": "update user endpoint",
	})
}

func getUserActivities(c *gin.Context) {
	userID := c.Param("userId")
	c.JSON(http.StatusOK, gin.H{
		"userId":  userID,
		"message": "get user activities endpoint",
	})
}
