package main

import (
	"finance/internal/handlers"
	"finance/internal/storage"
	"log"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	storage.InitDB()
	defer storage.CloseDB()

	// Set up Gin router
	r := gin.Default()

	// Find the path to the static files
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	staticPath := filepath.Join(filepath.Dir(basepath), "static")

	// Serve static files and main page
	r.Static("/static", staticPath)
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "index.html"))
	})

	// Serve authentication page
	r.GET("/auth", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "auth.html"))
	})

	// Public routes
	r.GET("/healthz", handlers.HealthCheck)
	r.POST("/auth/register", handlers.RegisterUser)
	r.POST("/auth/login", handlers.LoginUser)

	// Protected routes
	auth := r.Group("/")
	auth.Use(handlers.AuthMiddleware())
	{
		auth.POST("/auth/refresh", handlers.RefreshToken)

		// Deposit routes
		auth.POST("/deposit/create", handlers.CreateDeposit)
		auth.DELETE("/deposit/delete", handlers.DeleteDeposit)
		auth.POST("/deposit/transfer", handlers.TransferBetweenAccounts)
		auth.POST("/deposit/block", handlers.BlockDeposit)
		auth.POST("/deposit/unblock", handlers.UnblockDeposit)
		auth.POST("/deposit/freeze", handlers.FreezeDeposit)
	}

	// Start server
	log.Println("Starting server on :8080")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
