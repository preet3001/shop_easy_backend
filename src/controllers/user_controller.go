package controllers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/preet3001/shop_easy_backend/src/database"
	"github.com/preet3001/shop_easy_backend/src/helpers"
	"github.com/preet3001/shop_easy_backend/src/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
		return
	}

	validationErr := Validate.Struct(user)

	if validationErr != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
		return
	}

	count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if count > 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email already exist"})
		return
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

	_, insertError := UserCollection.InsertOne(ctx, user)
	if insertError != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user item not created"})
		return
	}
	defer cancel()
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"_id":           user.User_id,
		"token":         user.Token,
		"refresh_token": user.Refresh_token,
	})
}

func Login(c *gin.Context) {
	var ctx, cancel = context.WithTimeout(context.Background(),100*time.Second)
	var user models.User
	var foundUser models.User
	if err:= c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
	}

	err:= UserCollection.FindOne(ctx,bson.M{"email":user.Email}).Decode(&foundUser)
	defer cancel()
	if err!= nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":"email or password incorrect"})
		return
	}
	passwordIsValid,msg := VerifyPassword(*user.Password,*foundUser.Password)
	if !passwordIsValid {
		c.JSON(http.StatusBadRequest,gin.H{"error":msg})
		return
	}
	if foundUser.Email == nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":"user not found"})
		return
	}
	token,refreshToken,_  := helpers.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_type,foundUser.User_id) 
	helpers.UpdateAllTokens(token,refreshToken,foundUser.User_id)
	err = UserCollection.FindOne(ctx,bson.M{"user_id":foundUser.User_id}).Decode(&foundUser)
	if err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return
	}
	c.JSON(http.StatusOK,foundUser)
}

func GetUsers(c *gin.Context) {
	if err:= helpers.CheckUserType(c,"ADMIN");err != nil {
		c.JSON(http.StatusBadRequest,gin.H{"error":err.Error()})
		return
	}
	var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second)
	recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
	if err !=nil || recordPerPage <1  {
		recordPerPage = 10
	}
	page, err1 := strconv.Atoi(c.Query("page"))
	if err1 != nil || page<1 {
		page = 1
	}

	startIndex := (page -1 ) *recordPerPage
	// startIndex,err = strconv.Atoi(c.Query("startIndex"))
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	// }
	matchStage := bson.D{{Key: "$match",Value: bson.D{{}}}}
	groupStage := bson.D{{Key: "$group",Value: bson.D{
		{Key: "_id",Value: bson.D{{Key: "_id",Value: "null"}}},
		{Key: "total_count",Value: bson.D{{Key: "$sum",Value: 1}}},
		{Key: "data",Value: bson.D{{Key: "$push",Value: "$$ROOT"}}},
	}}}
	projectStage := bson.D{
		{Key: "$project",Value: bson.D{
			{Key: "_id",Value: 0},
			{Key: "total_count",Value: 1},
			{Key: "user_items",Value: bson.D{{Key: "$slice",Value: []interface{}{"$data",startIndex,recordPerPage}}}},
		}},
	}
	result, err := UserCollection.Aggregate(ctx,mongo.Pipeline{matchStage,groupStage,projectStage})
	defer cancel()
	if err != nil {
		c.JSON(http.StatusInternalServerError,gin.H{"error":err.Error()})
	}
	var allUsers []bson.M
	if err = result.All(ctx, &allUsers);err != nil {
		log.Fatal(err)
	}
	c.JSON(http.StatusOK,allUsers[0])
}

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
