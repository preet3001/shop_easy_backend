package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func mapUrls(){
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
		  "message": "pong",
		})
	  })
}