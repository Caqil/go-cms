package admin

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"go-cms/internal/database"
	"go-cms/internal/database/models"
	"go-cms/internal/plugins"
	"go-cms/internal/themes"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// UploadPlugin handles plugin zip file uploads with improved error handling
func (h *Handler) UploadPlugin(c *gin.Context) {
	var tempPath string

	// Cleanup function to ensure temp files are always removed
	defer func() {
		if tempPath != "" {
			if err := os.Remove(tempPath); err != nil && !os.IsNotExist(err) {
				log.Printf("Warning: Failed to remove temp file %s: %v", tempPath, err)
			}
		}
	}()

	log.Printf("[PLUGIN_UPLOAD] Starting plugin upload process")

	// Get uploaded file
	file, header, err := c.Request.FormFile("plugin")
	if err != nil {
		log.Printf("[PLUGIN_UPLOAD] Error getting form file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No plugin file provided"})
		return
	}
	defer file.Close()

	log.Printf("[PLUGIN_UPLOAD] Received file: %s, size: %d bytes", header.Filename, header.Size)

	// Validate file size (max 100MB)
	const maxFileSize = 100 << 20 // 100MB
	if header.Size > maxFileSize {
		log.Printf("[PLUGIN_UPLOAD] File too large: %d bytes (max: %d)", header.Size, maxFileSize)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum size is 100MB"})
		return
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		log.Printf("[PLUGIN_UPLOAD] Invalid file extension: %s", header.Filename)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only .zip files are allowed"})
		return
	}

	// Create temporary directory for upload
	tempDir := "./temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		log.Printf("[PLUGIN_UPLOAD] Failed to create temp directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp directory"})
		return
	}

	// Generate unique temp filename to avoid conflicts
	timestamp := time.Now().Unix()
	tempFilename := fmt.Sprintf("%d_%s", timestamp, header.Filename)
	tempPath = filepath.Join(tempDir, tempFilename)

	log.Printf("[PLUGIN_UPLOAD] Saving to temp path: %s", tempPath)

	// Save uploaded file temporarily
	dst, err := os.Create(tempPath)
	if err != nil {
		log.Printf("[PLUGIN_UPLOAD] Failed to create temp file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save uploaded file"})
		return
	}

	// Copy file contents
	bytesWritten, err := io.Copy(dst, file)
	dst.Close() // Close immediately after copy

	if err != nil {
		log.Printf("[PLUGIN_UPLOAD] Failed to copy file contents: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy uploaded file"})
		return
	}

	log.Printf("[PLUGIN_UPLOAD] Successfully saved %d bytes to temp file", bytesWritten)

	// Extract plugin name from filename (remove .zip extension)
	pluginName := strings.TrimSuffix(header.Filename, ".zip")

	// Validate plugin name
	if !isValidPluginName(pluginName) {
		log.Printf("[PLUGIN_UPLOAD] Invalid plugin name: %s", pluginName)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid plugin name. Use only lowercase letters, numbers, and hyphens",
		})
		return
	}

	log.Printf("[PLUGIN_UPLOAD] Plugin name: %s", pluginName)

	// Check if plugin already exists
	collection := h.db.Collection("plugins")
	var existingPlugin models.PluginMetadata
	err = collection.FindOne(context.Background(), bson.M{"name": pluginName}).Decode(&existingPlugin)
	if err == nil {
		log.Printf("[PLUGIN_UPLOAD] Plugin %s already exists, will update", pluginName)
	} else if err != mongo.ErrNoDocuments {
		log.Printf("[PLUGIN_UPLOAD] Database error checking existing plugin: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Validate the zip file
	log.Printf("[PLUGIN_UPLOAD] Validating plugin zip file")
	validationResult, err := h.pluginManager.ValidatePlugin(tempPath)
	if err != nil {
		log.Printf("[PLUGIN_UPLOAD] Plugin validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Plugin validation failed: %v", err),
		})
		return
	}

	if !validationResult.IsValid {
		log.Printf("[PLUGIN_UPLOAD] Plugin validation failed. Errors: %v, Warnings: %v",
			validationResult.Errors, validationResult.Warnings)
		c.JSON(http.StatusBadRequest, gin.H{
			"error":    "Plugin validation failed",
			"details":  validationResult.Errors,
			"warnings": validationResult.Warnings,
		})
		return
	}

	log.Printf("[PLUGIN_UPLOAD] Plugin validation successful")

	// If plugin exists and is active, deactivate it first
	if existingPlugin.Name != "" && existingPlugin.IsActive {
		log.Printf("[PLUGIN_UPLOAD] Deactivating existing plugin for update")
		if err := h.pluginManager.UnloadPlugin(pluginName); err != nil {
			log.Printf("[PLUGIN_UPLOAD] Warning: Failed to unload existing plugin: %v", err)
		}
	}

	// Install the plugin
	log.Printf("[PLUGIN_UPLOAD] Installing plugin")
	if err := h.pluginManager.InstallPluginFromZip(tempPath, pluginName); err != nil {
		log.Printf("[PLUGIN_UPLOAD] Plugin installation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Plugin installation failed: %v", err),
		})
		return
	}

	log.Printf("[PLUGIN_UPLOAD] Plugin installation successful")

	// Get plugin info for database storage
	pluginInfo, err := h.pluginManager.GetPluginInfo(pluginName)
	if err != nil {
		log.Printf("[PLUGIN_UPLOAD] Failed to get plugin info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Plugin installed but failed to get info",
		})
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

	// If plugin existed, preserve creation date
	if existingPlugin.Name != "" {
		pluginMetadata.CreatedAt = existingPlugin.CreatedAt
	}

	// Upsert the plugin metadata with proper error handling
	log.Printf("[PLUGIN_UPLOAD] Saving plugin metadata to database")
	filter := bson.M{"name": pluginInfo.Name}
	opts := options.Replace().SetUpsert(true)

	_, err = collection.ReplaceOne(context.Background(), filter, pluginMetadata, opts)
	if err != nil {
		log.Printf("[PLUGIN_UPLOAD] Database error saving plugin metadata: %v", err)
		// Try to clean up the installed plugin since DB save failed
		if unloadErr := h.pluginManager.UnloadPlugin(pluginName); unloadErr != nil {
			log.Printf("[PLUGIN_UPLOAD] Failed to cleanup plugin after DB error: %v", unloadErr)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Plugin installed but failed to save metadata",
		})
		return
	}

	log.Printf("[PLUGIN_UPLOAD] Plugin upload completed successfully: %s v%s",
		pluginInfo.Name, pluginInfo.Version)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":     "Plugin uploaded and installed successfully",
		"plugin_name": pluginInfo.Name,
		"filename":    header.Filename,
		"version":     pluginInfo.Version,
		"author":      pluginInfo.Author,
		"description": pluginInfo.Description,
	})
}

// Helper function to validate plugin names
func isValidPluginName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}

	// Only allow lowercase letters, numbers, and hyphens
	// Must start with a letter
	matched, _ := regexp.MatchString(`^[a-z][a-z0-9-]*$`, name)
	return matched
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
