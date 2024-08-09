package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func StateError(c *gin.Context, err error) {
	log.Println(err)
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": fmt.Sprintf("error: %v", err),
	})
}

func buildRoutes(server *gin.Engine, users map[string]string) {

	authorized := server.Group("/", gin.BasicAuth(users))

	server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authorized.Handle("LOCK", "/tfstates/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		state, err := DB.GetState(name)
		if err != nil {
			StateError(ctx, err)
		}
		if state == nil {
			state = NewState(name, DB)
			state.db.SaveNewState(state)
		} else if state.locked {
			ctx.JSON(http.StatusLocked, gin.H{
				"status": "already locked",
			})

		} else {
			state.Lock()
			ctx.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		}
	})

	authorized.Handle("UNLOCK", "/tfstates/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		state, err := DB.GetState(name)
		if err != nil {
			StateError(ctx, err)
		}
		if state == nil {
			state = NewState(name, DB)
			state.db.SaveNewState(state)
		} else if !state.locked {
			ctx.JSON(http.StatusConflict, gin.H{
				"status": "already unlocked",
			})

		} else {
			state.Unlock()
			ctx.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		}
	})

	authorized.Handle("GET", "/tfstates/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		state, err := DB.GetState(name)
		if err != nil {
			StateError(ctx, err)
		}
		if state == nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "state not found",
			})
		} else {
			ctx.String(http.StatusOK, state.contents)
		}
	})

	authorized.Handle("POST", "/tfstates/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		state, err := DB.GetState(name)
		if err != nil {
			StateError(ctx, err)
		}
		if state == nil {
			state = NewState(name, DB)
		}
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			StateError(ctx, err)
		}
		state.Update(string(data))
		ctx.String(http.StatusOK, "ok")
	})
}
