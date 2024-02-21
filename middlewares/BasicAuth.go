package middlewares

import (
	"os"

	"github.com/gin-gonic/gin"
)

func BasicAuth(c *gin.Context) {
	// reed OPERATOR_USERNAME and OPERATOR_PASSWORD from HEADER
	// if OPERATOR_USERNAME and OPERATOR_PASSWORD is not equal to .env file
	// then return 401
	// else continue
	username, password, ok := c.Request.BasicAuth()
	if !ok {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})


		c.Abort()
		return
	}

	if username != os.Getenv("BROADCAST_USERNAME") || password != os.Getenv("BROADCAST_PASSWORD") {
		c.JSON(401, gin.H{
			"message": "Unauthorized",
		})

		c.Abort()
		return
	}

	c.Next()

}
