package main

import (
	"github.com/gin-gonic/gin"
)

var MainServer *gin.Engine

func serve(users map[string]string) {
	MainServer = gin.Default()
	buildRoutes(MainServer, users)
	MainServer.Run("0.0.0.0:8081")
}
