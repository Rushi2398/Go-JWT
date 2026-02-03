package routes

import (
	controller "github.com/Rushi2398/Go-JWT/controllers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func AuthRoutes(incomingRoutes *gin.Engine, userCollection *mongo.Collection) {
	incomingRoutes.POST("/users/signup", controller.Signup(userCollection))
	incomingRoutes.POST("/users/login", controller.Login(userCollection))
}
