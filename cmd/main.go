package main

import (
	"finance/internal/handlers"
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
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "index.html"))
	})

	r.GET("/health", handlers.HealthCheck)
	r.POST("/deposit/create", handlers.CreateDeposit)
	r.POST("/deposit/transfer", handlers.TransferBetweenAccounts)
	r.POST("/deposit/block", handlers.BlockDeposit)
	r.POST("/deposit/unblock", handlers.UnblockDeposit)
	r.POST("/deposit/freeze", handlers.FreezeDeposit)
	r.DELETE("/deposit/delete", handlers.DeleteDeposit)

	r.Run(":8085")
}
