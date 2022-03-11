package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type helloworld struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

var ahelloworld = helloworld{ID: "test1", Message: "hello world!"}

func getHelloWorld(c *gin.Context) {
	c.JSON(http.StatusOK, ahelloworld)
}

func main() {
	router := gin.Default()
	router.GET("/helloworld", getHelloWorld)
	router.Run("0.0.0.0:8080")
}
