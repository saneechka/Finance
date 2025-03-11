package main

import (
	"finance/internal/handlers"
	"finance/internal/storage"
	"finance/internal/utils"
	"log"
	"path/filepath"
	"runtime"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize encryption
	if err := utils.InitEncryption(); err != nil {
		log.Printf("Warning: Failed to initialize encryption: %v", err)
	}

	// Initialize database
	storage.InitDB()
	defer storage.CloseDB()

	// Set up Gin router
	r := gin.Default()

	// Find the path to the static files
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	staticPath := filepath.Join(filepath.Dir(basepath), "static")

	// Serve static files and pages
	r.Static("/static", staticPath)
	r.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "index.html"))
	})
	r.GET("/deposits", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "deposits.html"))
	})
	r.GET("/loans", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "loans.html"))
	})

	// Serve authentication page
	r.GET("/auth", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "auth.html"))
	})

	// Serve admin page
	r.GET("/admin", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "admin.html"))
	})

	// Serve operator page
	r.GET("/operator", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "operator.html"))
	})

	// Serve manager page
	r.GET("/manager", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "manager.html"))
	})

	// Public routes
	r.GET("/health", handlers.HealthCheck)
	r.POST("/auth/register", handlers.RegisterUser)
	r.POST("/auth/login", handlers.LoginUser)

	// Protected routes with authentication middleware
	auth := r.Group("/")
	auth.Use(handlers.AuthMiddleware())
	{
		auth.POST("/auth/refresh", handlers.RefreshToken)

		// Admin routes for user approval
		auth.GET("/admin/pending-users", handlers.GetPendingUsers)
		auth.POST("/admin/approve-user", handlers.ApproveUser)
		auth.POST("/admin/reject-user", handlers.RejectUser)

		// New admin routes for logs and action cancellation
		auth.GET("/admin/action-logs", handlers.GetAllActionLogs)
		auth.POST("/admin/cancel-user-actions", handlers.CancelAllUserActions)

		// Operator routes
		auth.GET("/operator/statistics", handlers.GetTransactionStatistics)
		auth.GET("/operator/transactions", handlers.GetTransactions)
		auth.POST("/operator/cancel-transaction", handlers.CancelTransaction)
	}

	// Register deposit API endpoints
	depositRoutes := r.Group("/deposit")
	depositRoutes.Use(handlers.AuthMiddleware())
	{
		depositRoutes.POST("/create", handlers.CreateDeposit)
		depositRoutes.DELETE("/delete", handlers.DeleteDeposit)
		depositRoutes.POST("/transfer", handlers.TransferBetweenAccounts)
		depositRoutes.POST("/freeze", handlers.FreezeDeposit)
		depositRoutes.POST("/block", handlers.BlockDeposit)
		depositRoutes.POST("/unblock", handlers.UnblockDeposit)
		depositRoutes.GET("/list", handlers.GetDeposits)
	}

	// Register loan API endpoints
	loanRoutes := r.Group("/loan")
	loanRoutes.Use(handlers.AuthMiddleware())
	{
		loanRoutes.POST("/request", handlers.RequestLoan)
		loanRoutes.GET("/list", handlers.GetUserLoans)
		loanRoutes.GET("/:id", handlers.GetLoanDetails)
		loanRoutes.POST("/payment", handlers.MakeLoanPayment)
		loanRoutes.GET("/rates", handlers.GetLoanRates)
	}

	// Admin loan routes
	adminLoanRoutes := r.Group("/admin/loans")
	adminLoanRoutes.Use(handlers.AuthMiddleware())
	{
		adminLoanRoutes.GET("/pending", handlers.GetPendingLoans)
		adminLoanRoutes.POST("/approve", handlers.ApproveLoan)
		adminLoanRoutes.POST("/reject", handlers.RejectLoan)
	}

	// Manager routes
	managerRoutes := r.Group("/manager")
	managerRoutes.Use(handlers.AuthMiddleware())
	handlers.RegisterManagerRoutes(managerRoutes)

	// Start server
	log.Println("Starting server on :8081")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
