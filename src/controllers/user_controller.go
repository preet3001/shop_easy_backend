package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/preet3001/shop_easy_backend/src/database"
	"github.com/preet3001/shop_easy_backend/src/helpers"
	"github.com/preet3001/shop_easy_backend/src/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection = database.OpenCollection(database.Client, "user")
var Validate = validator.New()

func HashPassword(password string) string {
	byte, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(byte)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""
	if err != nil {
		msg = "email of password is incorrect"
		check = false
	}
	return check, msg
}

func SignUp(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var user models.User

	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	validationErr := Validate.Struct(user)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
	}

	count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if count > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email already exist"})
	}
	defer cancel()
	if err != nil {
		log.Panic(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error while checking email"})
		return
	}
	password := HashPassword(*user.Password)
	user.Password = &password

	count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
	defer cancel()

	if err != nil {
		log.Panic(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error checking email"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "phone already exist"})
		return
	}

	user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Local().Format(time.RFC3339))
	user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Local().Format(time.RFC3339))
	user.ID = primitive.NewObjectID()
	user.User_id = user.ID.Hex()
	token, refreshToken, err := helpers.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_type, user.User_id)
	if err != nil {
		log.Panic(err.Error())
	}
	user.Token = &token
	user.Refresh_token = &refreshToken

	resultInsertionNumber, insertError := UserCollection.InsertOne(ctx, user)
	if insertError != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user item not created"})
	}
	defer cancel()
	c.JSON(http.StatusOK, resultInsertionNumber)
}

func Login(c *gin.Context) {}

func GetUsers(c *gin.Context) {}

func GetUser(c *gin.Context) {
	userId := c.Param("user_id")
	if err := helpers.MatchUserTypeToUid(c, userId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	var user models.User
	err := UserCollection.FindOne(ctx, bson.M{"user_id": userId}).Decode(&user)
	defer cancel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}
