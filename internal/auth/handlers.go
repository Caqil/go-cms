package auth

import (
	"context"
	"net/http"
	"time"

	"go-cms/internal/database"
	"go-cms/internal/database/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Handler struct {
	db        *database.DB
	jwtSecret string
}

func NewHandler(db *database.DB, jwtSecret string) *Handler {
	return &Handler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// Register handles user registration
func (h *Handler) Register(c *gin.Context) {
	var req models.UserRegistration
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	collection := h.db.Collection("users")

	existingUser := models.User{}
	err := collection.FindOne(context.Background(), bson.M{
		"$or": []bson.M{
			{"email": req.Email},
			{"username": req.Username},
		},
	}).Decode(&existingUser)

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email or username already exists"})
		return
	} else if err != mongo.ErrNoDocuments {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Create new user
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  req.Password,
		Role:      "user",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Hash password
	if err := user.HashPassword(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Insert user
	result, err := collection.InsertOne(context.Background(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate tokens
	tokens, err := GenerateTokenPair(
		user.ID.Hex(),
		user.Username,
		user.Email,
		user.Role,
		h.jwtSecret,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"tokens": tokens,
	})
}

// Login handles user authentication
func (h *Handler) Login(c *gin.Context) {
	var req models.UserLogin
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	collection := h.db.Collection("users")
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	// Verify password
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login time
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"last_login_at": now,
			"updated_at":    now,
		},
	}
	collection.UpdateOne(context.Background(), bson.M{"_id": user.ID}, update)

	// Generate tokens
	tokens, err := GenerateTokenPair(
		user.ID.Hex(),
		user.Username,
		user.Email,
		user.Role,
		h.jwtSecret,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":       user.ID.Hex(),
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
		"tokens": tokens,
	})
}

// RefreshToken handles token refresh
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate refresh token
	claims, err := ValidateToken(req.RefreshToken, h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Get user from database to ensure they still exist and are active
	collection := h.db.Collection("users")
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)
	var user models.User
	err = collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	// Generate new tokens
	tokens, err := GenerateTokenPair(
		user.ID.Hex(),
		user.Username,
		user.Email,
		user.Role,
		h.jwtSecret,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  tokens,
	})
}

// GetProfile returns user profile information
func (h *Handler) GetProfile(c *gin.Context) {
	userContext, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Get full user data from database
	collection := h.db.Collection("users")
	userID, _ := primitive.ObjectIDFromHex(userContext.UserID)
	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":            user.ID.Hex(),
			"username":      user.Username,
			"email":         user.Email,
			"role":          user.Role,
			"is_active":     user.IsActive,
			"created_at":    user.CreatedAt,
			"updated_at":    user.UpdatedAt,
			"last_login_at": user.LastLoginAt,
		},
	})
}

// UpdateProfile updates user profile information
func (h *Handler) UpdateProfile(c *gin.Context) {
	userContext, exists := GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	var req struct {
		Username string `json:"username,omitempty"`
		Email    string `json:"email,omitempty"`
		Password string `json:"password,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := h.db.Collection("users")
	userID, _ := primitive.ObjectIDFromHex(userContext.UserID)

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	if req.Username != "" {
		// Check if username is already taken
		existingUser := models.User{}
		err := collection.FindOne(context.Background(), bson.M{
			"username": req.Username,
			"_id":      bson.M{"$ne": userID},
		}).Decode(&existingUser)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
		update["$set"].(bson.M)["username"] = req.Username
	}

	if req.Email != "" {
		// Check if email is already taken
		existingUser := models.User{}
		err := collection.FindOne(context.Background(), bson.M{
			"email": req.Email,
			"_id":   bson.M{"$ne": userID},
		}).Decode(&existingUser)
		if err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already taken"})
			return
		}
		update["$set"].(bson.M)["email"] = req.Email
	}

	if req.Password != "" {
		user := models.User{Password: req.Password}
		if err := user.HashPassword(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		update["$set"].(bson.M)["password"] = user.Password
	}

	// Update user
	_, err := collection.UpdateOne(context.Background(), bson.M{"_id": userID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
	})
}
