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

	// Serve jsauthentication page
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

	// Serve external specialist page
	r.GET("/external", func(c *gin.Context) {
		c.File(filepath.Join(staticPath, "external.html"))
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
	}

	// Admin routes - consolidated to prevent duplicates
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(handlers.AuthMiddleware())
	{
		// User management
		adminRoutes.GET("/pending-users", handlers.GetPendingUsers)
		adminRoutes.POST("/approve-user", handlers.ApproveUser)
		adminRoutes.POST("/reject-user", handlers.RejectUser)

		// Action logs
		adminRoutes.GET("/action-logs", handlers.GetAllActionLogs)
		adminRoutes.POST("/cancel-user-actions", handlers.CancelAllUserActions)

		// Loan management
		adminRoutes.GET("/loans/pending", handlers.GetPendingLoans)
		adminRoutes.POST("/loans/approve", handlers.ApproveLoan)
		adminRoutes.POST("/loans/reject", handlers.RejectLoan)

		// External specialist request management
		adminRoutes.GET("/external/pending-requests", handlers.GetPendingExternalRequests)
		adminRoutes.POST("/external/approve", handlers.ApproveExternalRequest)
		adminRoutes.POST("/external/reject", handlers.RejectExternalRequest)
	}

	// Operator routes
	operatorRoutes := r.Group("/operator")
	operatorRoutes.Use(handlers.AuthMiddleware())
	{
		operatorRoutes.GET("/statistics", handlers.GetTransactionStatistics)
		operatorRoutes.GET("/actions", handlers.GetUserActions)
		operatorRoutes.GET("/recent-actions", handlers.GetRecentActions)
		operatorRoutes.GET("/users", handlers.GetUsers)
		operatorRoutes.GET("/users/:id/last-action", handlers.GetUserLastAction)
		operatorRoutes.POST("/cancel-action", handlers.CancelLastOperation)
		operatorRoutes.GET("/transactions", handlers.GetTransactions)
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

	// Manager routes
	managerRoutes := r.Group("/manager")
	managerRoutes.Use(handlers.AuthMiddleware())
	handlers.RegisterManagerRoutes(managerRoutes)

	// External Enterprise Specialist routes
	externalRoutes := r.Group("/external")
	externalRoutes.Use(handlers.AuthMiddleware())
	{
		externalRoutes.POST("/salary-project", handlers.SubmitSalaryProject)
		externalRoutes.POST("/transfer-request", handlers.RequestEnterpriseTransfer)
		externalRoutes.GET("/transfers", handlers.GetEnterpriseTransfers)
		externalRoutes.GET("/salary-projects", handlers.GetSalaryProjects)
		externalRoutes.GET("/enterprises", handlers.GetUserEnterprises)
	}

	// Start server
	log.Println("Starting server on :8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
