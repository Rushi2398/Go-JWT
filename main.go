package main

import (
	"os"

	"github.com/Rushi2398/Go-JWT/database"
	routes "github.com/Rushi2398/Go-JWT/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	userCollection := database.OpenCollection(database.Client, "user")
	database.CreateUserIndexes(userCollection)

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router, userCollection)
	routes.UserRoutes(router, userCollection)

	router.GET("/api-1", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": "Access granted for api-1"})
	})

	router.GET("/api-2", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"success": "Access granted for api-2"})
	})

	router.Run(":" + port)
}
