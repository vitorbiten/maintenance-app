package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitorbiten/maintenance/api/app/auth"
)

func SetMiddlewareJSON(next http.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Content-Type", "application/json")
		c.Next()
	}
}

func SetMiddlewareAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Print("Hello, World!")
		err := auth.TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			fmt.Println(err)
			c.Abort()
			return
		}
		c.Next()
	}
}
