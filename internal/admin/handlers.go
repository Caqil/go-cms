package admin

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
}

func NewHandler(db *database.DB, pluginManager *plugins.Manager, themeManager *themes.Manager) *Handler {
	return &Handler{
		db:            db,
		pluginManager: pluginManager,
		themeManager:  themeManager,
	}
}

// Dashboard endpoint
func (h *Handler) GetDashboard(c *gin.Context) {
	// Get plugin statistics
	pluginCount := len(h.pluginManager.GetAllPlugins())

	// Get theme statistics
	themeCount := len(h.themeManager.GetAllThemes())

	// Get user count from database
	userCollection := h.db.Collection("users")
	userCount, _ := userCollection.CountDocuments(context.Background(), bson.M{})

	c.JSON(http.StatusOK, gin.H{
		"stats": gin.H{
			"plugins": pluginCount,
			"themes":  themeCount,
			"users":   userCount,
		},
		"recent_activity": []gin.H{}, // Can be extended with actual activity
	})
}

// Get admin menu with plugin contributions
func (h *Handler) GetMenu(c *gin.Context) {
	// Base admin menu items
	baseMenu := []plugins.AdminMenuItem{
		{
			ID:    "dashboard",
			Title: "Dashboard",
			Icon:  "dashboard",
			URL:   "/admin/dashboard",
			Order: 1,
		},
		{
			ID:    "plugins",
			Title: "Plugins",
			Icon:  "puzzle-piece",
			URL:   "/admin/plugins",
			Order: 2,
			Children: []plugins.AdminMenuItem{
				{
					ID:    "plugin-list",
					Title: "Installed Plugins",
					URL:   "/admin/plugins",
					Order: 1,
				},
				{
					ID:    "plugin-upload",
					Title: "Upload Plugin",
					URL:   "/admin/plugins/upload",
					Order: 2,
				},
			},
		},
		{
			ID:    "themes",
			Title: "Themes",
			Icon:  "palette",
			URL:   "/admin/themes",
			Order: 3,
		},
		{
			ID:    "settings",
			Title: "Settings",
			Icon:  "settings",
			URL:   "/admin/settings",
			Order: 100,
		},
	}

	// Get plugin menu items
	pluginMenuItems := h.pluginManager.GetAdminMenuItems()

	// Combine base menu with plugin menus
	allMenuItems := append(baseMenu, pluginMenuItems...)

	c.JSON(http.StatusOK, gin.H{
		"menu": allMenuItems,
	})
}

// Get all plugins with their settings
func (h *Handler) GetPlugins(c *gin.Context) {
	// Get loaded plugins
	loadedPlugins := h.pluginManager.GetAllPlugins()

	// Get plugin metadata from database
	collection := h.db.Collection("plugins")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plugins"})
		return
	}
	defer cursor.Close(context.Background())

	var dbPlugins []models.PluginMetadata
	if err := cursor.All(context.Background(), &dbPlugins); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode plugins"})
		return
	}

	// Combine loaded plugins with database metadata
	var response []gin.H
	for name, plugin := range loadedPlugins {
		info := plugin.GetInfo()
		settings := plugin.GetSettings()

		// Find corresponding database record
		var dbPlugin *models.PluginMetadata
		for _, p := range dbPlugins {
			if p.Name == name {
				dbPlugin = &p
				break
			}
		}

		pluginData := gin.H{
			"name":        info.Name,
			"version":     info.Version,
			"description": info.Description,
			"author":      info.Author,
			"website":     info.Website,
			"is_loaded":   true,
			"settings":    settings,
		}

		if dbPlugin != nil {
			pluginData["is_active"] = dbPlugin.IsActive
			pluginData["id"] = dbPlugin.ID.Hex()
		}

		response = append(response, pluginData)
	}

	c.JSON(http.StatusOK, gin.H{
		"plugins": response,
	})
}

// Upload and install plugin
func (h *Handler) UploadPlugin(c *gin.Context) {
	// Get uploaded file
	file, header, err := c.Request.FormFile("plugin")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Validate file extension
	if filepath.Ext(header.Filename) != ".so" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Only .so files are allowed"})
		return
	}

	// Create plugins directory if it doesn't exist
	pluginsDir := "./plugins"
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plugins directory"})
		return
	}

	// Save file
	pluginPath := filepath.Join(pluginsDir, header.Filename)
	dst, err := os.Create(pluginPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plugin file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to copy plugin file"})
		return
	}

	// Load the plugin
	if err := h.pluginManager.LoadPlugin(pluginPath); err != nil {
		// Remove the file if loading failed
		os.Remove(pluginPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to load plugin: %v", err)})
		return
	}

	// Get plugin info and save to database
	pluginName := filepath.Base(header.Filename[:len(header.Filename)-3]) // Remove .so extension
	if plugin, exists := h.pluginManager.GetPlugin(pluginName); exists {
		info := plugin.GetInfo()
		settings := plugin.GetSettings()

		pluginMetadata := models.PluginMetadata{
			Name:        info.Name,
			Version:     info.Version,
			Description: info.Description,
			Author:      info.Author,
			Website:     info.Website,
			Filename:    header.Filename,
			IsActive:    true,
			Settings:    convertToModelSettings(settings),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		collection := h.db.Collection("plugins")
		_, err := collection.InsertOne(context.Background(), pluginMetadata)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save plugin metadata"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Plugin uploaded and installed successfully",
		"filename": header.Filename,
	})
}

// Toggle plugin activation
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
		h.pluginManager.UnloadPlugin(pluginName)
	} else {
		// If activating, reload the plugin
		pluginPath := filepath.Join("./plugins", plugin.Filename)
		if err := h.pluginManager.LoadPlugin(pluginPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load plugin"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Plugin status updated",
		"is_active": newStatus,
	})
}

// Get plugin settings
func (h *Handler) GetPluginSettings(c *gin.Context) {
	pluginName := c.Param("name")

	collection := h.db.Collection("plugins")
	var plugin models.PluginMetadata
	err := collection.FindOne(context.Background(), bson.M{"name": pluginName}).Decode(&plugin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plugin":   plugin.Name,
		"settings": plugin.Settings,
	})
}

// Update plugin settings
func (h *Handler) UpdatePluginSettings(c *gin.Context) {
	pluginName := c.Param("name")

	var settingsUpdate map[string]interface{}
	if err := c.ShouldBindJSON(&settingsUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	collection := h.db.Collection("plugins")

	// Get current plugin settings
	var plugin models.PluginMetadata
	err := collection.FindOne(context.Background(), bson.M{"name": pluginName}).Decode(&plugin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// Update settings values
	for i, setting := range plugin.Settings {
		if newValue, exists := settingsUpdate[setting.Key]; exists {
			plugin.Settings[i].Value = newValue
		}
	}

	// Save to database
	update := bson.M{
		"$set": bson.M{
			"settings":   plugin.Settings,
			"updated_at": time.Now(),
		},
	}

	_, err = collection.UpdateOne(context.Background(), bson.M{"name": pluginName}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Settings updated successfully",
		"settings": plugin.Settings,
	})
}

// Delete plugin
func (h *Handler) DeletePlugin(c *gin.Context) {
	pluginName := c.Param("name")

	// Unload plugin first
	h.pluginManager.UnloadPlugin(pluginName)

	collection := h.db.Collection("plugins")

	// Get plugin metadata to find filename
	var plugin models.PluginMetadata
	err := collection.FindOne(context.Background(), bson.M{"name": pluginName}).Decode(&plugin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plugin not found"})
		return
	}

	// Delete file
	pluginPath := filepath.Join("./plugins", plugin.Filename)
	if err := os.Remove(pluginPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plugin file"})
		return
	}

	// Delete from database
	_, err = collection.DeleteOne(context.Background(), bson.M{"name": pluginName})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plugin metadata"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Plugin deleted successfully",
	})
}

// Helper function to convert plugin settings
func convertToModelSettings(pluginSettings []plugins.PluginSetting) []models.PluginSetting {
	var modelSettings []models.PluginSetting
	for _, setting := range pluginSettings {
		modelSettings = append(modelSettings, models.PluginSetting{
			Key:         setting.Key,
			Label:       setting.Label,
			Type:        setting.Type,
			Value:       setting.Value,
			Description: setting.Description,
			Options:     setting.Options,
			Required:    setting.Required,
		})
	}
	return modelSettings
}
