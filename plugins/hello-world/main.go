package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HelloWorldPlugin implements the Plugin interface
type HelloWorldPlugin struct {
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
	Type        string      `json:"type"` // text, number, boolean, select
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
	Options     []string    `json:"options,omitempty"` // For select type
	Required    bool        `json:"required"`
}

// Message represents a simple message structure
type Message struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Timestamp time.Time `json:"timestamp"`
}

// NewPlugin is the entry point that will be called by the plugin manager
func NewPlugin() Plugin {
	return &HelloWorldPlugin{}
}

func (p *HelloWorldPlugin) GetInfo() PluginInfo {
	return PluginInfo{
		Name:        "hello-world",
		Version:     "1.0.0",
		Description: "A simple Hello World plugin demonstrating basic CMS plugin functionality",
		Author:      "CMS Developer",
		Website:     "https://example.com",
	}
}

func (p *HelloWorldPlugin) Initialize(deps *PluginDependencies) error {
	p.deps = deps
	p.setDefaultSettings()

	// Perform any initialization here
	// You could set up database connections, load configuration, etc.
	return nil
}

func (p *HelloWorldPlugin) setDefaultSettings() {
	p.settings = map[string]interface{}{
		"enabled":         true,
		"welcome_message": "Hello, World! Welcome to our CMS!",
		"show_timestamp":  true,
		"max_messages":    10,
		"theme_color":     "blue",
	}
}

func (p *HelloWorldPlugin) RegisterRoutes(router *gin.RouterGroup) {
	// Basic API routes
	router.GET("/hello", p.handleHello)
	router.GET("/info", p.handleInfo)
	router.GET("/status", p.handleStatus)

	// Message management routes
	router.GET("/messages", p.handleGetMessages)
	router.POST("/messages", p.handleCreateMessage)
	router.GET("/messages/:id", p.handleGetMessage)
	router.DELETE("/messages/:id", p.handleDeleteMessage)

	// Settings routes
	router.GET("/settings", p.handleGetSettings)
	router.PUT("/settings", p.handleUpdateSettings)

	// Health check
	router.GET("/health", p.handleHealthCheck)
}

func (p *HelloWorldPlugin) GetAdminMenuItems() []AdminMenuItem {
	return []AdminMenuItem{
		{
			ID:    "hello-world-menu",
			Title: "Hello World",
			Icon:  "globe",
			URL:   "/admin/plugins/hello-world",
			Order: 60,
			Children: []AdminMenuItem{
				{
					ID:    "hello-world-dashboard",
					Title: "Dashboard",
					URL:   "/admin/plugins/hello-world/dashboard",
					Order: 1,
				},
				{
					ID:    "hello-world-messages",
					Title: "Messages",
					URL:   "/admin/plugins/hello-world/messages",
					Order: 2,
				},
				{
					ID:    "hello-world-settings",
					Title: "Settings",
					URL:   "/admin/plugins/hello-world/settings",
					Order: 3,
				},
			},
		},
	}
}

func (p *HelloWorldPlugin) GetSettings() []PluginSetting {
	return []PluginSetting{
		{
			Key:         "enabled",
			Label:       "Enable Hello World Plugin",
			Type:        "boolean",
			Value:       p.settings["enabled"],
			Description: "Enable or disable the Hello World plugin functionality",
			Required:    false,
		},
		{
			Key:         "welcome_message",
			Label:       "Welcome Message",
			Type:        "text",
			Value:       p.settings["welcome_message"],
			Description: "The message displayed to users",
			Required:    true,
		},
		{
			Key:         "show_timestamp",
			Label:       "Show Timestamps",
			Type:        "boolean",
			Value:       p.settings["show_timestamp"],
			Description: "Whether to display timestamps with messages",
			Required:    false,
		},
		{
			Key:         "max_messages",
			Label:       "Maximum Messages",
			Type:        "number",
			Value:       p.settings["max_messages"],
			Description: "Maximum number of messages to store",
			Required:    false,
		},
		{
			Key:         "theme_color",
			Label:       "Theme Color",
			Type:        "select",
			Value:       p.settings["theme_color"],
			Description: "Color theme for the plugin interface",
			Options:     []string{"blue", "green", "red", "purple", "orange"},
			Required:    false,
		},
	}
}

func (p *HelloWorldPlugin) Shutdown() error {
	// Perform cleanup here - close connections, save state, etc.
	return nil
}

// HTTP Handlers

func (p *HelloWorldPlugin) handleHello(c *gin.Context) {
	enabled, ok := p.settings["enabled"].(bool)
	if !ok || !enabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Hello World plugin is disabled",
		})
		return
	}

	message := p.settings["welcome_message"].(string)
	showTimestamp := p.settings["show_timestamp"].(bool)

	response := gin.H{
		"message": message,
		"plugin":  p.GetInfo().Name,
		"version": p.GetInfo().Version,
	}

	if showTimestamp {
		response["timestamp"] = time.Now().Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, response)
}

func (p *HelloWorldPlugin) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"info":     p.GetInfo(),
		"settings": p.GetSettings(),
		"status":   "active",
	})
}

func (p *HelloWorldPlugin) handleStatus(c *gin.Context) {
	enabled := p.settings["enabled"].(bool)

	c.JSON(http.StatusOK, gin.H{
		"plugin_name": p.GetInfo().Name,
		"version":     p.GetInfo().Version,
		"enabled":     enabled,
		"uptime":      time.Now().Format(time.RFC3339),
		"endpoints": []string{
			"/hello",
			"/info",
			"/status",
			"/messages",
			"/settings",
			"/health",
		},
	})
}

func (p *HelloWorldPlugin) handleGetMessages(c *gin.Context) {
	// In a real implementation, this would query the database
	// For now, we'll return mock data
	messages := []Message{
		{
			ID:        1,
			Content:   "Welcome to the Hello World plugin!",
			Author:    "System",
			Timestamp: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        2,
			Content:   "This is a sample message from the plugin",
			Author:    "Admin",
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    len(messages),
	})
}

func (p *HelloWorldPlugin) handleCreateMessage(c *gin.Context) {
	var requestData struct {
		Content string `json:"content" binding:"required"`
		Author  string `json:"author" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// In a real implementation, this would save to the database
	message := Message{
		ID:        int(time.Now().Unix()), // Simple ID generation
		Content:   requestData.Content,
		Author:    requestData.Author,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Message created successfully",
		"data":    message,
	})
}

func (p *HelloWorldPlugin) handleGetMessage(c *gin.Context) {
	id := c.Param("id")

	// In a real implementation, this would query the database
	message := Message{
		ID:        1,
		Content:   "Sample message content for ID: " + id,
		Author:    "System",
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

func (p *HelloWorldPlugin) handleDeleteMessage(c *gin.Context) {
	id := c.Param("id")

	// In a real implementation, this would delete from the database
	c.JSON(http.StatusOK, gin.H{
		"message": "Message deleted successfully",
		"id":      id,
	})
}

func (p *HelloWorldPlugin) handleGetSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"settings": p.settings,
	})
}

func (p *HelloWorldPlugin) handleUpdateSettings(c *gin.Context) {
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

func (p *HelloWorldPlugin) handleHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"plugin":    p.GetInfo().Name,
		"version":   p.GetInfo().Version,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
