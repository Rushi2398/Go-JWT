package routes

import (
	controller "github.com/Rushi2398/Go-JWT/controllers"
	middleware "github.com/Rushi2398/Go-JWT/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func UserRoutes(incomingRoutes *gin.Engine, userCollection *mongo.Collection) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers(userCollection))
	incomingRoutes.GET("/users/:user_id", controller.GetUser(userCollection))
}
