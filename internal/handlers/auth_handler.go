package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "strings"
	"time"

	"finance/internal/models"
	db "finance/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey []byte

func init() {
	// Get JWT secret key from environment variable or use default
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "your_secret_key" // Default key for development
		log.Println("Warning: Using default JWT secret key. Set JWT_SECRET_KEY environment variable in production.")
	}
	jwtKey = []byte(secretKey)
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func RegisterUser(c *gin.Context) {
	var userInput struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
		FullName string `json:"fullName"`
		Role     string `json:"role"`
	}

	if err := c.ShouldBindJSON(&userInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log registration attempt for debugging
	log.Printf("Registration attempt: username=%s, email=%s, role=%s",
		userInput.Username, userInput.Email, userInput.Role)

	// Validate username and password
	if userInput.Username == "" || userInput.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	// Set default role if not provided
	if userInput.Role == "" {
		userInput.Role = "client"
	}

	// Validate role is either 'client', 'admin', 'operator', or 'manager'
	if userInput.Role != "client" && userInput.Role != "admin" && userInput.Role != "operator" && userInput.Role != "manager" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role must be 'client', 'manager', 'admin', or 'operator'"})
		return
	}

	// Create user model from input
	user := &models.User{
		Username: userInput.Username,
		Email:    userInput.Email,
		Role:     userInput.Role,
		// Auto-approve admin, operator, and manager accounts, client accounts need approval
		Approved: userInput.Role == "admin" || userInput.Role == "operator" || userInput.Role == "manager",
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// Save user to database
	if err := db.SaveUser(user); err != nil {
		if err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		} else {
			log.Printf("Error registering user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user: " + err.Error()})
		}
		return
	}

	// Don't return the password in response
	user.Password = ""

	// Return different message based on role
	var message string
	if user.Role == "admin" {
		message = "Administrator registration successful. You can now log in."
	} else if user.Role == "operator" {
		message = "Operator registration successful. You can now log in."
	} else if user.Role == "manager" {
		message = "Manager registration successful. You can now log in."
	} else {
		message = "Registration successful. Your account is pending approval by an administrator."
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":    user,
		"message": message,
	})
}

func LoginUser(c *gin.Context) {
	var loginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from database
	user, err := db.GetUserByUsername(loginRequest.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate user"})
		}
		return
	}

	// Compare password with stored hash
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginRequest.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// Check if the user is approved
	if !user.Approved {
		c.JSON(http.StatusForbidden, gin.H{"error": "your account is pending approval by an administrator"})
		return
	}

	// if !

	// Create token expiring in 24 hours
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	// Add role and approval status to the response
	c.JSON(http.StatusOK, gin.H{
		"token":    tokenString,
		"expires":  expirationTime,
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"approved": user.Approved,
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		// Check if the header has the correct format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract the token from the Authorization header
		// Format should be "Bearer {token}"
		tokenString := authHeader[7:] // Skip "Bearer "

		// Parse and validate the token
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token signature"})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			}
			c.Abort()
			return
		}

		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Add user ID to the context - ensure it's stored as an int
		c.Set("userID", int(claims.UserID))
		c.Next()
	}
}

func RefreshToken(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userIDValue, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Handle type conversion safely
	var userID int
	switch v := userIDValue.(type) {
	case int:
		userID = v
	case float64:
		userID = int(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID type"})
		return
	}

	// Create new token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Get user info for subject
	user, err := db.GetUserByID(userID)
	if err == nil {
		claims.Subject = user.Username
	}



	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   tokenString,
		"expires": expirationTime,
		"user_id": userID,
	})
}

// Added new handler for admin to get pending users
func GetPendingUsers(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := db.IsUserAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	// Get pending users
	pendingUsers, err := db.GetPendingUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve pending users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_users": pendingUsers})
}

// Added new handler for admin to approve users
func ApproveUser(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := db.IsUserAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	var request struct {
		UserID int `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.ApproveUser(request.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user approved successfully"})
}

// Added new handler for admin to reject users
func RejectUser(c *gin.Context) {
	userID, exists := getUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Check if user is admin
	isAdmin, err := db.IsUserAdmin(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify admin status"})
		return
	}

	if !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "admin privileges required"})
		return
	}

	var request struct {
		UserID int `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.RejectUser(request.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user rejected successfully"})
}

// Check if user has privileges to access certain functionality
func hasRole(userID int, roles ...string) bool {
	user, err := db.GetUserByID(userID)
	if err != nil {
		return false
	}

	for _, role := range roles {
		if user.Role == role {
			return true
		}
	}
	return false
}

// getUserID extracts the user ID from the Gin context

