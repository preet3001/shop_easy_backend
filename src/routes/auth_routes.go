package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/preet3001/shop_easy_backend/src/controllers"
)

func AuthRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.POST("users/signup",controllers.SignUp)
	incomingRoutes.POST("users/login",controllers.Login)
}