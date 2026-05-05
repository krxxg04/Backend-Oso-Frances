package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Mateo Leche es gay y se la come",
		})
	})

	r.Run(":8080")
}