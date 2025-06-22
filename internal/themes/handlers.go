package themes

import (
	"net/http"
	"path/filepath"

	"go-cms/internal/auth"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) *Handler {
	return &Handler{
		manager: manager,
	}
}

// GetAll returns all available themes
func (h *Handler) GetAll(c *gin.Context) {
	themes := h.manager.GetAllThemes()

	var themeList []gin.H
	for _, theme := range themes {
		themeData := gin.H{
			"name":             theme.Name,
			"version":          theme.Version,
			"description":      theme.Description,
			"author":           theme.Author,
			"author_uri":       theme.AuthorURI,
			"screenshot":       theme.Screenshot,
			"tags":             theme.Tags,
			"min_version":      theme.MinVersion,
			"required_plugins": theme.RequiredPlugins,
			"is_active":        theme.IsActive,
			"installed_at":     theme.InstalledAt,
			"updated_at":       theme.UpdatedAt,
		}

		themeList = append(themeList, themeData)
	}

	c.JSON(http.StatusOK, gin.H{
		"themes": themeList,
		"active": h.manager.active,
	})
}

// GetTheme returns a specific theme
func (h *Handler) GetTheme(c *gin.Context) {
	themeName := c.Param("name")

	theme, exists := h.manager.GetTheme(themeName)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Theme not found"})
		return
	}

	// Get theme assets
	assets, err := h.manager.GetThemeAssets(themeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get theme assets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"theme": gin.H{
			"name":             theme.Name,
			"version":          theme.Version,
			"description":      theme.Description,
			"author":           theme.Author,
			"author_uri":       theme.AuthorURI,
			"screenshot":       theme.Screenshot,
			"tags":             theme.Tags,
			"min_version":      theme.MinVersion,
			"required_plugins": theme.RequiredPlugins,
			"templates":        theme.Templates,
			"customization":    theme.Customization,
			"is_active":        theme.IsActive,
			"installed_at":     theme.InstalledAt,
			"updated_at":       theme.UpdatedAt,
			"assets":           assets,
		},
	})
}

// ActivateTheme activates a specific theme
func (h *Handler) ActivateTheme(c *gin.Context) {
	themeName := c.Param("name")

	// Get user context
	userContext, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Check if user has permission (admin or super admin)
	if userContext.Role != "admin" && userContext.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	// Activate theme
	err := h.manager.SetActiveTheme(themeName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Theme activated successfully",
		"active_theme": themeName,
	})
}

// GetCustomization returns theme customization settings
func (h *Handler) GetCustomization(c *gin.Context) {
	themeName := c.Param("name")

	customization, err := h.manager.GetThemeCustomization(themeName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"theme":         themeName,
		"customization": customization,
	})
}

// UpdateCustomization updates theme customization settings
func (h *Handler) UpdateCustomization(c *gin.Context) {
	themeName := c.Param("name")

	// Get user context
	userContext, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Check permissions
	if userContext.Role != "admin" && userContext.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	var customization Customization
	if err := c.ShouldBindJSON(&customization); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update customization
	err := h.manager.UpdateThemeCustomization(themeName, customization)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Theme customization updated successfully",
		"theme":         themeName,
		"customization": customization,
	})
}

// InstallTheme handles theme installation
func (h *Handler) InstallTheme(c *gin.Context) {
	// Get user context
	userContext, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Check permissions (only super admin can install themes)
	if userContext.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("theme")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No theme file uploaded"})
		return
	}
	defer file.Close()

	// Validate file (should be a zip file)
	if filepath.Ext(header.Filename) != ".zip" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme must be a zip file"})
		return
	}

	// In a real implementation, you would:
	// 1. Extract the zip file
	// 2. Validate theme structure
	// 3. Install the theme

	c.JSON(http.StatusOK, gin.H{
		"message":  "Theme installation would be implemented here",
		"filename": header.Filename,
	})
}

// UninstallTheme handles theme removal
func (h *Handler) UninstallTheme(c *gin.Context) {
	themeName := c.Param("name")

	// Get user context
	userContext, exists := auth.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	// Check permissions
	if userContext.Role != "super_admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
		return
	}

	// Uninstall theme
	err := h.manager.UninstallTheme(themeName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Theme uninstalled successfully",
		"theme":   themeName,
	})
}
