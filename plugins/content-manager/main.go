package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ContentManagerPlugin demonstrates a more complex plugin with database integration
type ContentManagerPlugin struct {
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

type ContentItem struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	Author    string             `bson:"author" json:"author"`
	Status    string             `bson:"status" json:"status"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// NewPlugin is the entry point that will be called by the plugin manager
func NewPlugin() Plugin {
	return &ContentManagerPlugin{
		settings: make(map[string]interface{}),
	}
}

func (p *ContentManagerPlugin) GetInfo() PluginInfo {
	return PluginInfo{
		Name:        "content-manager",
		Version:     "1.2.0",
		Description: "Advanced content management system with configurable features",
		Author:      "CMS Team",
		Website:     "https://example.com/content-manager",
	}
}

func (p *ContentManagerPlugin) Initialize(deps *PluginDependencies) error {
	p.deps = deps

	// Load settings from database
	if err := p.loadSettings(); err != nil {
		// Use default settings if loading fails
		p.setDefaultSettings()
	}

	return nil
}

func (p *ContentManagerPlugin) loadSettings() error {
	// In a real implementation, you would load settings from the database
	// For this example, we'll simulate loading settings
	db := p.deps.Database
	if db == nil {
		return nil
	}

	// Load settings from the plugins collection
	// This would typically involve querying the database for this plugin's settings
	p.settings = map[string]interface{}{
		"posts_per_page":   10,
		"allow_comments":   true,
		"moderation_mode":  "auto",
		"content_types":    []string{"post", "page", "article"},
		"auto_save":        true,
		"max_content_size": 50000,
	}

	return nil
}

func (p *ContentManagerPlugin) setDefaultSettings() {
	p.settings = map[string]interface{}{
		"posts_per_page":   5,
		"allow_comments":   false,
		"moderation_mode":  "manual",
		"content_types":    []string{"post", "page"},
		"auto_save":        false,
		"max_content_size": 10000,
	}
}

func (p *ContentManagerPlugin) RegisterRoutes(router *gin.RouterGroup) {
	// Content CRUD routes
	router.GET("/content", p.handleListContent)
	router.POST("/content", p.handleCreateContent)
	router.GET("/content/:id", p.handleGetContent)
	router.PUT("/content/:id", p.handleUpdateContent)
	router.DELETE("/content/:id", p.handleDeleteContent)

	// Plugin-specific routes
	router.GET("/stats", p.handleContentStats)
	router.GET("/export", p.handleExportContent)
	router.POST("/import", p.handleImportContent)
}

func (p *ContentManagerPlugin) GetAdminMenuItems() []AdminMenuItem {
	return []AdminMenuItem{
		{
			ID:    "content-manager-menu",
			Title: "Content Manager",
			Icon:  "edit",
			URL:   "/admin/plugins/content-manager",
			Order: 50,
			Children: []AdminMenuItem{
				{
					ID:    "content-list",
					Title: "All Content",
					URL:   "/admin/plugins/content-manager/content",
					Order: 1,
				},
				{
					ID:    "content-create",
					Title: "Create New",
					URL:   "/admin/plugins/content-manager/create",
					Order: 2,
				},
				{
					ID:    "content-stats",
					Title: "Statistics",
					URL:   "/admin/plugins/content-manager/stats",
					Order: 3,
				},
				{
					ID:    "content-settings",
					Title: "Settings",
					URL:   "/admin/plugins/content-manager/settings",
					Order: 4,
				},
			},
		},
	}
}

func (p *ContentManagerPlugin) GetSettings() []PluginSetting {
	return []PluginSetting{
		{
			Key:         "posts_per_page",
			Label:       "Posts Per Page",
			Type:        "number",
			Value:       p.settings["posts_per_page"],
			Description: "Number of posts to display per page",
			Required:    true,
		},
		{
			Key:         "allow_comments",
			Label:       "Allow Comments",
			Type:        "boolean",
			Value:       p.settings["allow_comments"],
			Description: "Enable or disable comments on content",
			Required:    false,
		},
		{
			Key:         "moderation_mode",
			Label:       "Comment Moderation",
			Type:        "select",
			Value:       p.settings["moderation_mode"],
			Description: "How to handle comment moderation",
			Options:     []string{"auto", "manual", "disabled"},
			Required:    true,
		},
		{
			Key:         "auto_save",
			Label:       "Auto Save",
			Type:        "boolean",
			Value:       p.settings["auto_save"],
			Description: "Automatically save content while editing",
			Required:    false,
		},
		{
			Key:         "max_content_size",
			Label:       "Max Content Size (characters)",
			Type:        "number",
			Value:       p.settings["max_content_size"],
			Description: "Maximum number of characters allowed in content",
			Required:    true,
		},
	}
}

func (p *ContentManagerPlugin) Shutdown() error {
	// Perform any necessary cleanup
	return nil
}

// HTTP Handlers

func (p *ContentManagerPlugin) handleListContent(c *gin.Context) {
	// Get pagination settings from plugin settings
	perPage := p.settings["posts_per_page"].(int)

	// Mock content list (in real implementation, query database)
	content := []ContentItem{
		{
			ID:        primitive.NewObjectID(),
			Title:     "Sample Post 1",
			Content:   "This is sample content...",
			Author:    "admin",
			Status:    "published",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        primitive.NewObjectID(),
			Title:     "Sample Post 2",
			Content:   "This is another sample...",
			Author:    "admin",
			Status:    "draft",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"content":      content,
		"per_page":     perPage,
		"total_pages":  1,
		"current_page": 1,
		"plugin_info":  p.GetInfo(),
	})
}

func (p *ContentManagerPlugin) handleCreateContent(c *gin.Context) {
	var newContent ContentItem
	if err := c.ShouldBindJSON(&newContent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check content size limit
	maxSize := p.settings["max_content_size"].(int)
	if len(newContent.Content) > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Content exceeds maximum size limit",
			"limit": maxSize,
		})
		return
	}

	// Set timestamps
	newContent.ID = primitive.NewObjectID()
	newContent.CreatedAt = time.Now()
	newContent.UpdatedAt = time.Now()

	// In real implementation, save to database
	c.JSON(http.StatusCreated, gin.H{
		"message": "Content created successfully",
		"content": newContent,
	})
}

func (p *ContentManagerPlugin) handleGetContent(c *gin.Context) {
	contentID := c.Param("id")

	// Mock content retrieval
	content := ContentItem{
		ID:        primitive.NewObjectID(),
		Title:     "Sample Content",
		Content:   "This is the full content...",
		Author:    "admin",
		Status:    "published",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"content":     content,
		"content_id":  contentID,
		"plugin_info": p.GetInfo(),
	})
}

func (p *ContentManagerPlugin) handleUpdateContent(c *gin.Context) {
	contentID := c.Param("id")

	var updates ContentItem
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check content size limit
	maxSize := p.settings["max_content_size"].(int)
	if len(updates.Content) > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Content exceeds maximum size limit",
			"limit": maxSize,
		})
		return
	}

	// Set update timestamp
	updates.UpdatedAt = time.Now()

	c.JSON(http.StatusOK, gin.H{
		"message":    "Content updated successfully",
		"content_id": contentID,
		"content":    updates,
	})
}

func (p *ContentManagerPlugin) handleDeleteContent(c *gin.Context) {
	contentID := c.Param("id")

	// In real implementation, delete from database
	c.JSON(http.StatusOK, gin.H{
		"message":    "Content deleted successfully",
		"content_id": contentID,
	})
}

func (p *ContentManagerPlugin) handleContentStats(c *gin.Context) {
	// Mock statistics based on plugin settings
	allowComments := p.settings["allow_comments"].(bool)

	stats := gin.H{
		"total_content":      42,
		"published":          38,
		"drafts":             4,
		"total_comments":     0,
		"pending_moderation": 0,
		"plugin_settings": gin.H{
			"comments_enabled": allowComments,
			"moderation_mode":  p.settings["moderation_mode"],
			"auto_save":        p.settings["auto_save"],
		},
	}

	if allowComments {
		stats["total_comments"] = 156
		stats["pending_moderation"] = 3
	}

	c.JSON(http.StatusOK, stats)
}

func (p *ContentManagerPlugin) handleExportContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Export functionality would be implemented here",
		"format":  "JSON",
		"plugin":  p.GetInfo().Name,
	})
}

func (p *ContentManagerPlugin) handleImportContent(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Import functionality would be implemented here",
		"plugin":  p.GetInfo().Name,
	})
}
