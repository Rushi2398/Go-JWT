package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Rushi2398/Go-JWT/database"
	routes "github.com/Rushi2398/Go-JWT/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	// ------Mongo Connection---------
	client, err := database.ConnectMongo(os.Getenv("MONGODB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	// Ensure DB Disconnect
	defer func() {
		cleanCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := client.Disconnect(cleanCtx); err != nil {
			log.Println("error disconnecting mongo:", err)
		} else {
			log.Println("✅ MongoDB disconnected")
		}
	}()

	userCollection := database.OpenCollection(client, os.Getenv("MONGODB_NAME"), "user")

	if err := database.CreateUserIndexes(ctx, userCollection); err != nil {
		log.Fatal("failed to create mongo indexes:", err)
	}

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
