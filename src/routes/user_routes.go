package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/preet3001/shop_easy_backend/src/controllers"
	"github.com/preet3001/shop_easy_backend/src/middlewares"
)


func UserRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.Use(middlewares.Authenticate)
	incomingRoutes.GET("/users",controllers.GetUsers)
	incomingRoutes.GET("/users/:user_id",controllers.GetUser)
}