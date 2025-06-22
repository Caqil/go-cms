package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTMiddleware validates JWT tokens for protected routes
func JWTMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract token
		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := ValidateToken(tokenString, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// AdminRequired middleware ensures user has admin role
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role type"})
			c.Abort()
			return
		}

		if userRole != "admin" && userRole != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SuperAdminRequired middleware ensures user has super admin role
func SuperAdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			c.Abort()
			return
		}

		userRole, ok := role.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid role type"})
			c.Abort()
			return
		}

		if userRole != "super_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetUserFromContext extracts user information from gin context
func GetUserFromContext(c *gin.Context) (UserContext, bool) {
	userID, exists1 := c.Get("user_id")
	username, exists2 := c.Get("username")
	email, exists3 := c.Get("email")
	role, exists4 := c.Get("role")

	if !exists1 || !exists2 || !exists3 || !exists4 {
		return UserContext{}, false
	}

	return UserContext{
		UserID:   userID.(string),
		Username: username.(string),
		Email:    email.(string),
		Role:     role.(string),
	}, true
}

type UserContext struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
