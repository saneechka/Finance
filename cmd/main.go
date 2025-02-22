package main

import (
	"finance/internal/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/health", handlers.HealthCheck)
	r.POST("/deposit/create", handlers.CreateDeposit)

	r.Run(":8085")
}
