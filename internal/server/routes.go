package server

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kamuridesu/tf-backend-go/internal/db"
)

func StateError(c *gin.Context, err error) {
	slog.Error(fmt.Sprintf("got error: %s\n", err))
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": fmt.Sprintf("error: %v", err),
	})
}

func BuildRoutes(server *gin.Engine, database db.Database, users map[string]string) {

	authorized := server.Group("/", gin.BasicAuth(users))

	server.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	authorized.Handle("LOCK", "/tfstates/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		state, err := database.GetState(name)
		if err != nil {
			StateError(ctx, err)
			return
		}
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			slog.Error(err.Error())
		}
		slog.Info(string(data))
		if state == nil {
			state = db.NewState(name, database)
			err := state.Database.SaveNewState(state)
			if err != nil {
				StateError(ctx, err)
				return
			}
		} else if state.Locked {
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
		state, err := database.GetState(name)
		if err != nil {
			StateError(ctx, err)
			return
		}
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			slog.Error(err.Error())
		}
		slog.Info(string(data))
		if state == nil {
			state = db.NewState(name, database)
			err := state.Database.SaveNewState(state)
			if err != nil {
				StateError(ctx, err)
			}
		} else if !state.Locked {
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
		state, err := database.GetState(name)
		if err != nil {
			StateError(ctx, err)
			return
		}
		if state == nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"status": "state not found",
			})
		} else {
			ctx.String(http.StatusOK, state.Contents)
		}
	})

	authorized.Handle("POST", "/tfstates/:name", func(ctx *gin.Context) {
		name := ctx.Param("name")
		state, err := database.GetState(name)
		if err != nil {
			StateError(ctx, err)
			return
		}
		if state == nil {
			state = db.NewState(name, database)
			err := state.Database.SaveNewState(state)
			if err != nil {
				StateError(ctx, err)
				return
			}
		}
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			StateError(ctx, err)
			return
		}
		err = state.Update(string(data))
		if err != nil {
			StateError(ctx, err)
			return
		} else {
			ctx.String(http.StatusOK, "ok")
		}

	})
}
