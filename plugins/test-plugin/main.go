package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TestPluginPlugin implements the Plugin interface
type TestPluginPlugin struct {
	deps     *PluginDependencies
	settings map[string]interface{}
}

type PluginDependencies struct {
	Database interface{}
	Config   interface{}
}

type Plugin interface {
	GetInfo() PluginInfo
	Initialize(deps *PluginDependencies) error
	RegisterRoutes(router *gin.RouterGroup)
	GetAdminMenuItems() []AdminMenuItem
	GetSettings() []PluginSetting
	Shutdown() error
}

type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Website     string `json:"website,omitempty"`
}

type AdminMenuItem struct {
	ID       string          `json:"id"`
	Title    string          `json:"title"`
	Icon     string          `json:"icon,omitempty"`
	URL      string          `json:"url"`
	Parent   string          `json:"parent,omitempty"`
	Order    int             `json:"order"`
	Children []AdminMenuItem `json:"children,omitempty"`
}

type PluginSetting struct {
	Key         string      `json:"key"`
	Label       string      `json:"label"`
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	Options     []string    `json:"options,omitempty"`
	Required    bool        `json:"required"`
}

// NewPlugin is the entry point that will be called by the plugin manager
func NewPlugin() Plugin {
	return &TestPluginPlugin{
		settings: make(map[string]interface{}),
	}
}

func (p *TestPluginPlugin) GetInfo() PluginInfo {
	return PluginInfo{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Description: "A test-plugin plugin for the CMS",
		Author:      "Plugin Developer",
		Website:     "",
	}
}

func (p *TestPluginPlugin) Initialize(deps *PluginDependencies) error {
	p.deps = deps
	p.setDefaultSettings()
	
	// Perform any initialization here
	// You could set up database connections, load configuration, etc.
	return nil
}

func (p *TestPluginPlugin) setDefaultSettings() {
	p.settings = map[string]interface{}{
		"enabled":      true,
		"auto_update":  false,
		"cache_ttl":    3600,
		"debug_mode":   false,
	}
}

func (p *TestPluginPlugin) RegisterRoutes(router *gin.RouterGroup) {
	// Register plugin-specific routes
	router.GET("/", p.handleIndex)
	router.GET("/info", p.handleInfo)
	router.GET("/status", p.handleStatus)
	router.POST("/action", p.handleAction)
	
	// Admin routes (if needed)
	adminGroup := router.Group("/admin")
	{
		adminGroup.GET("/dashboard", p.handleAdminDashboard)
		adminGroup.GET("/settings", p.handleAdminSettings)
		adminGroup.PUT("/settings", p.handleUpdateSettings)
	}
}

func (p *TestPluginPlugin) GetAdminMenuItems() []AdminMenuItem {
	return []AdminMenuItem{
		{
			ID:    "test-plugin-menu",
			Title: "Test Plugin",
			Icon:  "puzzle-piece",
			URL:   "/admin/plugins/test-plugin",
			Order: 50,
			Children: []AdminMenuItem{
				{
					ID:    "test-plugin-dashboard",
					Title: "Dashboard",
					URL:   "/admin/plugins/test-plugin/admin/dashboard",
					Order: 1,
				},
				{
					ID:    "test-plugin-settings",
					Title: "Settings",
					URL:   "/admin/plugins/test-plugin/admin/settings",
					Order: 2,
				},
			},
		},
	}
}

func (p *TestPluginPlugin) GetSettings() []PluginSetting {
	return []PluginSetting{
		{
			Key:         "enabled",
			Label:       "Enable Test Plugin",
			Type:        "boolean",
			Value:       p.settings["enabled"],
			Description: "Enable or disable Test Plugin functionality",
			Required:    false,
		},
		{
			Key:         "auto_update",
			Label:       "Auto Update",
			Type:        "boolean",
			Value:       p.settings["auto_update"],
			Description: "Automatically update plugin when new version is available",
			Required:    false,
		},
		{
			Key:         "cache_ttl",
			Label:       "Cache TTL (seconds)",
			Type:        "number",
			Value:       p.settings["cache_ttl"],
			Description: "Time to live for cached data in seconds",
			Required:    true,
		},
		{
			Key:         "debug_mode",
			Label:       "Debug Mode",
			Type:        "boolean",
			Value:       p.settings["debug_mode"],
			Description: "Enable debug logging for this plugin",
			Required:    false,
		},
	}
}

func (p *TestPluginPlugin) Shutdown() error {
	// Perform cleanup here
	return nil
}

// HTTP Handlers

func (p *TestPluginPlugin) handleIndex(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugin":     p.GetInfo().Name,
		"version":    p.GetInfo().Version,
		"message":    "Welcome to Test Plugin!",
		"endpoints": []string{
			"/",
			"/info",
			"/status",
			"/action",
			"/admin/dashboard",
			"/admin/settings",
		},
	})
}

func (p *TestPluginPlugin) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"info":     p.GetInfo(),
		"settings": p.GetSettings(),
	})
}

func (p *TestPluginPlugin) handleStatus(c *gin.Context) {
	enabled := p.settings["enabled"].(bool)
	
	c.JSON(http.StatusOK, gin.H{
		"plugin":    p.GetInfo().Name,
		"version":   p.GetInfo().Version,
		"enabled":   enabled,
		"uptime":    time.Now().Format(time.RFC3339),
		"status":    "healthy",
	})
}

func (p *TestPluginPlugin) handleAction(c *gin.Context) {
	var requestData struct {
		Action string                 `json:"action"` binding:"required"`
		Data   map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Handle different actions
	switch requestData.Action {
	case "ping":
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"plugin":  p.GetInfo().Name,
		})
	case "test":
		c.JSON(http.StatusOK, gin.H{
			"message": "Test action executed successfully",
			"data":    requestData.Data,
		})
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Unknown action: " + requestData.Action,
		})
	}
}

func (p *TestPluginPlugin) handleAdminDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugin":  p.GetInfo().Name,
		"title":   "Test Plugin Dashboard",
		"stats": map[string]interface{}{
			"enabled":    p.settings["enabled"],
			"version":    p.GetInfo().Version,
			"cache_ttl":  p.settings["cache_ttl"],
			"debug_mode": p.settings["debug_mode"],
		},
	})
}

func (p *TestPluginPlugin) handleAdminSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugin":   p.GetInfo().Name,
		"settings": p.settings,
	})
}

func (p *TestPluginPlugin) handleUpdateSettings(c *gin.Context) {
	var newSettings map[string]interface{}

	if err := c.ShouldBindJSON(&newSettings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update settings (in a real implementation, you'd validate and persist these)
	for key, value := range newSettings {
		if _, exists := p.settings[key]; exists {
			p.settings[key] = value
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Settings updated successfully",
		"settings": p.settings,
	})
}
