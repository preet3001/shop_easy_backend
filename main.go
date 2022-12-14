package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/preet3001/shop_easy_backend/src/routes"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}
	router := gin.New()
	router.Use(gin.Logger())
	routes.AuthRoutes(router)
	routes.UserRoutes(router)
	router.GET("/api-1", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"sucess": "Access Granted for api 1",
		})
	})
	router.GET("/api-2", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"sucess": "Access Granted for api 2",
		})
	})
	router.Run(":" + port)
}
