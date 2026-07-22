package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler"
	"github.com/pocket7878/mono-repo-template-2026-summer/backend/internal/handler/api"
)

func NewRouter(todoHandler *handler.TodoHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://127.0.0.1:5173")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}
		c.Next()
	})
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api.RegisterHandlersWithOptions(router, todoHandler, api.GinServerOptions{
		ErrorHandler: func(c *gin.Context, err error, statusCode int) {
			c.Header("Content-Type", "application/problem+json")
			c.JSON(statusCode, api.Problem{
				Type:   "https://reference-todo.dev/problems/invalid-request",
				Title:  "Invalid request",
				Status: int32(statusCode),
			})
		},
	})
	return router
}
