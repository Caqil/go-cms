package admin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-cms/internal/database"
	"go-cms/internal/database/models"
	"go-cms/internal/plugins"
	"go-cms/internal/themes"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Handler struct {
	db            *database.DB
	pluginManager *plugins.Manager
	themeManager  *themes.Manager
	dashboard     *DashboardManager
}

func NewHandler(db *database.DB, pluginManager *plugins.Manager, themeManager *themes.Manager) *Handler {
	return &Handler{
		db:            db,
		pluginManager: pluginManager,
		themeManager:  themeManager,
		dashboard:     NewDashboardManager(db, pluginManager, themeManager),
	}
}

// GetDashboard returns dashboard statistics
func (h *Handler) GetDashboard(c *gin.Context) {
	dashboardData, err := h.dashboard.GetDashboardData()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard data"})
		return
	}

	c.JSON(http.StatusOK, dashboardData)
}

// GetMenu returns the admin menu structure
func (h *Handler) GetMenu(c *gin.Context) {
	menuManager := NewMenuManager(h.pluginManager)
	menu := menuManager.GetFullMenu()

	c.JSON(http.StatusOK, gin.H{
		"menu": menu,
	})
}

// GetPlugins returns list of all plugins
func (h *Handler) GetPlugins(c *gin.Context) {
	// Get loaded plugins
	loadedPlugins := h.pluginManager.GetAllPlugins()

	// Get plugin metadata from database
	collection := h.db.Collection("plugins")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plugins from database"})
		return
	}
	defer cursor.Close(context.Background())

	var dbPlugins []models.PluginMetadata
	if err := cursor.All(context.Background(), &dbPlugins); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode plugins"})
		return
	}

	// Merge loaded plugins with database metadata
	var responsePlugins []map[string]interface{}

	for _, dbPlugin := range dbPlugins {
		pluginData := map[string]interface{}{
			"name":        dbPlugin.Name,
			"version":     dbPlugin.Version,
			"description": dbPlugin.Description,
			"author":      dbPlugin.Author,
			"website":     dbPlugin.Website,
			"is_active":   dbPlugin.IsActive,
			"created_at":  dbPlugin.CreatedAt,
			"updated_at":  dbPlugin.UpdatedAt,
			"is_loaded":   false,
		}

		// Check if plugin is currently loaded
		if plugin, exists := loadedPlugins[dbPlugin.Name]; exists {
			pluginData["is_loaded"] = true
			pluginData["info"] = plugin.GetInfo()
			pluginData["settings"] = plugin.GetSettings()
		}

		responsePlugins = append(responsePlugins, pluginData)
	}

	// Get system information
	systemInfo, err := h.pluginManager.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plugins":     responsePlugins,
		"system_info": systemInfo,
	})
}

// UploadPlugin handles plugin zip file uploads
func (h *Handler) UploadPlugin(c *gin.Context) {
	// Get uploaded file
	file, header, err := c.Request.FormFile("plugin")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No plugin file provided"})
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .zip files are allowed"})
		return
	}

	// Create temporary directory for upload
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp directory"})
		return
	}

	// Save uploaded file temporarily
	tempPath := filepath.Join(tempDir, header.Filename)
	dst, err := os.Create(tempPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy uploaded file"})
		return
	}

	// Extract plugin name from filename (remove .zip extension)
	pluginName := strings.TrimSuffix(header.Filename, ".zip")

	// Validate plugin name
	if !isValidPluginName(pluginName) {
		os.Remove(tempPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plugin name. Use only lowercase letters, numbers, and hyphens"})
		return
	}

	// Validate the zip file
	validationResult, err := h.pluginManager.ValidatePlugin(tempPath)
	if err != nil {
		os.Remove(tempPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Plugin validation failed: %v", err)})
		return
	}

	if !validationResult.IsValid {
		os.Remove(tempPath)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Plugin validation failed",
			"details":  validationResult.Errors,
			"warnings": validationResult.Warnings,
		})
		return
	}

	// Install the plugin
	if err := h.pluginManager.InstallPluginFromZip(tempPath, pluginName); err != nil {
		os.Remove(tempPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Plugin installation failed: %v", err)})
		return
	}

	// Get plugin info for database storage
	pluginInfo, err := h.pluginManager.GetPluginInfo(pluginName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Plugin installed but failed to get info"})
		return
	}

	// Get plugin instance for settings
	var settings []models.PluginSetting
	if plugin, exists := h.pluginManager.GetPlugin(pluginInfo.Name); exists {
		pluginSettings := plugin.GetSettings()
		settings = convertToModelSettings(pluginSettings)
	}

	// Save plugin metadata to database
	pluginMetadata := models.PluginMetadata{
		Name:        pluginInfo.Name,
		Version:     pluginInfo.Version,
		Description: pluginInfo.Description,
		Author:      pluginInfo.Author,
		Website:     pluginInfo.Website,
		Filename:    header.Filename,
		IsActive:    true,
		Settings:    settings,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	collection := h.db.Collection("plugins")

	// Upsert the plugin metadata
	filter := bson.M{"name": pluginInfo.Name}
	_, err = collection.ReplaceOne(context.Background(), filter, pluginMetadata)
	if err != nil {
		_, err = collection.InsertOne(context.Background(), pluginMetadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plugin metadata"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Plugin uploaded and installed successfully",
		"plugin_name": pluginInfo.Name,
		"filename":    header.Filename,
		"version":     pluginInfo.Version,
	})
}

// TogglePlugin activates/deactivates a plugin
func (h *Handler) TogglePlugin(c *gin.Context) {
	pluginName := c.Param("name")

	collection := h.db.Collection("plugins")

	// Find plugin in database
	var plugin models.PluginMetadata
	err := collection.FindOne(context.Background(), bson.M{"name": pluginName}).Decode(&plugin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// Toggle active status
	newStatus := !plugin.IsActive
	update := bson.M{
		"$set": bson.M{
			"is_active":  newStatus,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"name": pluginName}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plugin status"})
		return
	}

	// If deactivating, unload the plugin
	if !newStatus {
		if err := h.pluginManager.UnloadPlugin(pluginName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unload plugin"})
			return
		}
	} else {
		// If activating, load the plugin
		if err := h.pluginManager.LoadPlugin(pluginName); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load plugin"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Plugin status updated",
		"is_active": newStatus,
	})
}

// DeletePlugin removes a plugin completely
func (h *Handler) DeletePlugin(c *gin.Context) {
	pluginName := c.Param("name")

	// Remove from database
	collection := h.db.Collection("plugins")
	_, err := collection.DeleteOne(context.Background(), bson.M{"name": pluginName})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove plugin from database"})
		return
	}

	// Uninstall the plugin (removes files and unloads)
	if err := h.pluginManager.UninstallPlugin(pluginName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to uninstall plugin: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin deleted successfully",
	})
}

// ReloadPlugin recompiles and reloads a plugin
func (h *Handler) ReloadPlugin(c *gin.Context) {
	pluginName := c.Param("name")

	if err := h.pluginManager.ReloadPlugin(pluginName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to reload plugin: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin reloaded successfully",
	})
}

// GetPluginSettings returns settings for a specific plugin
func (h *Handler) GetPluginSettings(c *gin.Context) {
	pluginName := c.Param("name")

	settings, err := h.pluginManager.GetPluginSettings(pluginName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plugin":   pluginName,
		"settings": settings,
	})
}

// UpdatePluginSettings updates settings for a specific plugin
func (h *Handler) UpdatePluginSettings(c *gin.Context) {
	pluginName := c.Param("name")

	var newSettings map[string]interface{}
	if err := c.ShouldBindJSON(&newSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current plugin settings
	currentSettings, err := h.pluginManager.GetPluginSettings(pluginName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Validate and update settings
	updatedSettings := make([]plugins.PluginSetting, len(currentSettings))
	copy(updatedSettings, currentSettings)

	for i, setting := range updatedSettings {
		if newValue, exists := newSettings[setting.Key]; exists {
			updatedSettings[i].Value = newValue
		}
	}

	// Save to database
	collection := h.db.Collection("plugins")
	update := bson.M{
		"$set": bson.M{
			"settings":   convertToModelSettings(updatedSettings),
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"name": pluginName}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Settings updated successfully",
		"settings": updatedSettings,
	})
}

// GetSystemInfo returns system information for plugin development
func (h *Handler) GetSystemInfo(c *gin.Context) {
	systemInfo, err := h.pluginManager.GetSystemInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get system info"})
		return
	}

	c.JSON(http.StatusOK, systemInfo)
}

// CleanupCache removes old compiled plugin files
func (h *Handler) CleanupCache(c *gin.Context) {
	// Default to 7 days
	maxAge := 7 * 24 * time.Hour

	if err := h.pluginManager.CleanupCache(maxAge); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cleanup cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleaned up successfully",
	})
}

// HotReloadAll reloads all plugins without restarting the server
func (h *Handler) HotReloadAll(c *gin.Context) {
	if err := h.pluginManager.HotReload(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Hot reload failed: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "All plugins reloaded successfully",
	})
}

// Utility functions

func isValidPluginName(name string) bool {
	if len(name) == 0 || len(name) > 50 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	return true
}

func convertToModelSettings(pluginSettings []plugins.PluginSetting) []models.PluginSetting {
	modelSettings := make([]models.PluginSetting, len(pluginSettings))
	for i, setting := range pluginSettings {
		modelSettings[i] = models.PluginSetting{
			Key:         setting.Key,
			Label:       setting.Label,
			Type:        setting.Type,
			Value:       setting.Value,
			Description: setting.Description,
			Options:     setting.Options,
			Required:    setting.Required,
		}
	}
	return modelSettings
}
