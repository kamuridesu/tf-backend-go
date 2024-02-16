package main

import (
	"github.com/gin-gonic/gin"
)

var server *gin.Engine

func serve() {
	server = gin.Default()
	buildRoutes()
	server.Run("0.0.0.0:8080")
}
