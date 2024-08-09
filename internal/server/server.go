package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kamuridesu/tf-backend-go/internal/db"
	"github.com/kamuridesu/tf-backend-go/internal/routes"
)

var MainServer *gin.Engine

func Serve(users map[string]string, database *db.Database) {
	MainServer = gin.Default()
	routes.BuildRoutes(MainServer, database, users)
	MainServer.Run("0.0.0.0:8081")
}
