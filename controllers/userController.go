package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	helper "github.com/Rushi2398/Go-JWT/helpers"
	model "github.com/Rushi2398/Go-JWT/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func HashPassword(userPassword *string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*userPassword), 14)
	if err != nil {
		log.Fatal(err)
	}
	return string(hashedPassword)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	if err != nil {
		return false, "email or password is incorrect"
	}
	return true, ""
}

func Signup(userCollection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user model.User
		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(user)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(c, bson.M{"email": user.Email})
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for email"})
			return
		}

		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		count, err = userCollection.CountDocuments(c, bson.M{"phone": user.Phone})
		if err != nil {
			log.Println(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for phone number"})
			return
		}

		if count > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "phone number already exists"})
			return
		}

		password := HashPassword(user.Password)
		user.Password = &password

		user.Created_at = time.Now()
		user.Updated_at = time.Now()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		token, refreshToken, err := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}
		user.Token = &token
		user.Refresh_token = &refreshToken
		resultInsertionNumber, insertErr := userCollection.InsertOne(c, user)
		if insertErr != nil {
			msg := "User item was not created"
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		ctx.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login(userCollection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user model.User
		var foundUser model.User

		if err := ctx.ShouldBindJSON(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := userCollection.FindOne(c, bson.M{"email": user.Email}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "email or password is incorrect"})
			return
		}
		passwordIsValid, msg := VerifyPassword(*foundUser.Password, *user.Password)
		if !passwordIsValid {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		if foundUser.Email == nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
		token, refreshToken, err := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type, foundUser.User_id)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
			return
		}

		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id, userCollection)
		err = userCollection.FindOne(c, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		foundUser.Password = nil
		foundUser.Refresh_token = nil
		ctx.JSON(http.StatusOK, foundUser)
	}
}

func GetUser(userCollection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")
		if err := helper.MatchUserTypeToUid(ctx, userId); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var c, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user model.User
		err := userCollection.FindOne(c, bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user.Password = nil
		user.Refresh_token = nil
		ctx.JSON(http.StatusOK, user)
	}
}

func GetUsers(userCollection *mongo.Collection) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if err := helper.CheckUserType(ctx, "ADMIN"); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		recordPerPage, err := strconv.Atoi(ctx.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, err := strconv.Atoi(ctx.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}
		startIndex, err := strconv.Atoi(ctx.Query("startIndex"))
		if err != nil {
			startIndex = (page - 1) * recordPerPage
		}
		matchStage := bson.D{
			{
				Key: "$match", Value: bson.D{},
			},
		}
		groupStage := bson.D{
			{
				Key: "$group", Value: bson.D{
					{
						Key: "_id", Value: bson.D{
							{
								Key: "id", Value: "null",
							},
							{
								Key: "total_count", Value: bson.D{
									{
										Key: "$sum", Value: 1,
									},
								},
							},
							{
								Key: "data", Value: bson.D{
									{
										Key: "$push", Value: "$$ROOT",
									},
								},
							},
						},
					},
				},
			},
		}
		projectStage := bson.D{
			{
				Key: "$project", Value: bson.D{
					{Key: "_id", Value: 0},
					{Key: "total_count", Value: 1},
					{
						Key: "user_items",
						Value: bson.D{
							{
								Key: "$map",
								Value: bson.D{
									{
										Key: "input",
										Value: bson.D{
											{
												Key:   "$slice",
												Value: []interface{}{"$data", startIndex, recordPerPage},
											},
										},
									},
									{
										Key: "as", Value: "user",
									},
									{
										Key: "in",
										Value: bson.D{
											{Key: "user_id", Value: "$$user.user_id"},
											{Key: "email", Value: "$$user.email"},
											{Key: "first_name", Value: "$$user.first_name"},
											{Key: "last_name", Value: "$$user.last_name"},
											{Key: "phone", Value: "$$user.phone"},
											{Key: "user_type", Value: "$$user.user_type"},
											{Key: "created_at", Value: "$$user.created_at"},
											{Key: "updated_at", Value: "$$user.updated_at"},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		result, err := userCollection.Aggregate(c, mongo.Pipeline{
			matchStage, groupStage, projectStage,
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing all user items"})
		}
		var allUsers []bson.M
		if err := result.All(c, &allUsers); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, allUsers[0])
	}
}
