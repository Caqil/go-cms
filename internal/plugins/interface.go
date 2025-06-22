package plugins

import (
	"github.com/gin-gonic/gin"
)

// Plugin represents the interface that all plugins must implement
type Plugin interface {
	// GetInfo returns basic plugin information
	GetInfo() PluginInfo

	// Initialize sets up the plugin with necessary dependencies
	Initialize(deps *PluginDependencies) error

	// RegisterRoutes allows plugins to register their own API routes
	RegisterRoutes(router *gin.RouterGroup)

	// GetAdminMenuItems returns menu items for the admin dashboard
	GetAdminMenuItems() []AdminMenuItem

	// GetSettings returns configurable settings for the plugin
	GetSettings() []PluginSetting

	// Shutdown performs cleanup when the plugin is unloaded
	Shutdown() error
}

type PluginInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Author      string `json:"author"`
	Website     string `json:"website,omitempty"`
}

type PluginDependencies struct {
	Database interface{} // Will be *database.DB
	Config   interface{} // Will be *config.Config
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
