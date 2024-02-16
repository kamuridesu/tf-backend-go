package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func buildRoutes() {
	server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	server.Handle("LOCK", "/tfstates/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		fmt.Println(name)
		ctx.Status(200)
	})
}
