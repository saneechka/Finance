package main

import (
	"finance/internal/handlers"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Get the path to the static directory
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	staticPath := filepath.Join(filepath.Dir(basepath), "static")

	// Serve static files
	r.Static("/static", staticPath)

	// Auth routes (public)
	r.GET("/auth", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "auth.html"))
	})

	// Auth endpoints

	// Main application route with auth check
	r.GET("/", func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.Redirect(http.StatusTemporaryRedirect, "/auth")
			return
		}
		c.File(filepath.Join(staticPath, "index.html"))
	})

	// Protected API routes
	api := r.Group("/api")
	// Removed authentication middleware
	{
		api.GET("/health", handlers.HealthCheck)

		// Add new route for requests page with no auth check
		api.GET("/deposit/requests", func(c *gin.Context) {
			c.File(filepath.Join(staticPath, "index.html"))
		})

		// Deposit operations
		deposits := api.Group("/deposit")
		{
			deposits.POST("/create", handlers.CreateDeposit)
			deposits.POST("/transfer", handlers.TransferBetweenAccounts)
			deposits.POST("/block", handlers.BlockDeposit)
			deposits.POST("/unblock", handlers.UnblockDeposit)
			deposits.POST("/freeze", handlers.FreezeDeposit)
			deposits.DELETE("/delete", handlers.DeleteDeposit)
		}
	}

	// Handle 404 and auth redirects
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 404 handling
		if len(path) >= 4 && path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API endpoint not found",
			})
			return
		}

		// Auth check for other routes
		token := c.GetHeader("Authorization")
		if token == "" {
			c.Redirect(http.StatusTemporaryRedirect, "/auth")
			return
		}

		// Serve index.html for client-side routing
		c.File(filepath.Join(staticPath, "index.html"))
	})

	r.Run(":8085")
}
