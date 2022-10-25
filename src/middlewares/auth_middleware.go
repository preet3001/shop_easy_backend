package middlewares

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/preet3001/shop_easy_backend/src/helpers"
)

func Authenticate(c *gin.Context) {
	clientToken := c.Request.Header.Get("Authorization")
	if clientToken == "" {
		c.JSON(http.StatusBadRequest,gin.H{"error":"authorization not provided"})
		c.Abort()
		return
	}

	claims,err := helpers.ValidateToken(clientToken)
	if err != "" {
		c.JSON(http.StatusBadRequest,gin.H{"error":err})
		c.Abort()
		return
	}
	c.Set("email",claims.Email)
	c.Set("first_nama",claims.First_name)
	c.Set("last_name",claims.Last_name)
	c.Set("uid",claims.Uid)
	c.Set("user_type",claims.User_type)
	c.Next()
}
