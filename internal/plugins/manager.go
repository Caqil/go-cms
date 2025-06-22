package plugins

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"plugin"
	"strings"

	"go-cms/internal/database"
	"go-cms/internal/database/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type Manager struct {
	plugins     map[string]Plugin
	pluginPaths map[string]string
	deps        *PluginDependencies
	db          *database.DB
}

func NewManager() *Manager {
	return &Manager{
		plugins:     make(map[string]Plugin),
		pluginPaths: make(map[string]string),
	}
}

func (m *Manager) SetDependencies(deps *PluginDependencies) {
	m.deps = deps
	if db, ok := deps.Database.(*database.DB); ok {
		m.db = db
	}
}

func (m *Manager) LoadPlugins(pluginDir string) error {
	// Load plugins from database that should be active
	if m.db != nil {
		if err := m.loadActivePluginsFromDB(pluginDir); err != nil {
			log.Printf("Warning: Failed to load plugins from database: %v", err)
		}
	}

	// Also scan for any .so files that might not be in database yet
	soFiles, err := filepath.Glob(filepath.Join(pluginDir, "*.so"))
	if err != nil {
		return fmt.Errorf("failed to find plugin files: %w", err)
	}

	for _, soFile := range soFiles {
		pluginName := strings.TrimSuffix(filepath.Base(soFile), ".so")
		if _, exists := m.plugins[pluginName]; !exists {
			if err := m.LoadPlugin(soFile); err != nil {
				log.Printf("Failed to load plugin %s: %v", soFile, err)
			}
		}
	}

	return nil
}

func (m *Manager) loadActivePluginsFromDB(pluginDir string) error {
	collection := m.db.Collection("plugins")
	cursor, err := collection.Find(context.Background(), bson.M{"is_active": true})
	if err != nil {
		return err
	}
	defer cursor.Close(context.Background())

	var activePlugins []models.PluginMetadata
	if err := cursor.All(context.Background(), &activePlugins); err != nil {
		return err
	}

	for _, pluginMeta := range activePlugins {
		pluginPath := filepath.Join(pluginDir, pluginMeta.Filename)
		if err := m.LoadPlugin(pluginPath); err != nil {
			log.Printf("Failed to load active plugin %s: %v", pluginMeta.Name, err)
			// Mark as inactive in database if loading fails
			m.markPluginInactive(pluginMeta.Name)
		}
	}

	return nil
}

func (m *Manager) markPluginInactive(pluginName string) {
	if m.db == nil {
		return
	}

	collection := m.db.Collection("plugins")
	update := bson.M{
		"$set": bson.M{
			"is_active": false,
		},
	}

	_, err := collection.UpdateOne(context.Background(), bson.M{"name": pluginName}, update)
	if err != nil {
		log.Printf("Failed to mark plugin %s as inactive: %v", pluginName, err)
	}
}

func (m *Manager) LoadPlugin(path string) error {
	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin: %w", err)
	}

	// Look for the NewPlugin symbol
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin must export 'NewPlugin' function: %w", err)
	}

	// Cast to the expected function type
	newPluginFunc, ok := symbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("NewPlugin must be a function that returns Plugin interface")
	}

	// Create plugin instance
	pluginInstance := newPluginFunc()
	info := pluginInstance.GetInfo()

	// Check if plugin is already loaded
	if _, exists := m.plugins[info.Name]; exists {
		return fmt.Errorf("plugin %s is already loaded", info.Name)
	}

	// Initialize the plugin
	if m.deps != nil {
		if err := pluginInstance.Initialize(m.deps); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", info.Name, err)
		}
	}

	// Store the plugin
	m.plugins[info.Name] = pluginInstance
	m.pluginPaths[info.Name] = path

	log.Printf("Loaded plugin: %s v%s", info.Name, info.Version)
	return nil
}

func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	plugin, exists := m.plugins[name]
	return plugin, exists
}

func (m *Manager) GetAllPlugins() map[string]Plugin {
	return m.plugins
}

func (m *Manager) RegisterRoutes(router *gin.RouterGroup) {
	// Create a plugin group
	pluginGroup := router.Group("/plugins")

	for name, plugin := range m.plugins {
		// Create a sub-group for each plugin
		pluginRouter := pluginGroup.Group("/" + strings.ToLower(name))
		plugin.RegisterRoutes(pluginRouter)
	}
}

func (m *Manager) GetAdminMenuItems() []AdminMenuItem {
	var allItems []AdminMenuItem

	for _, plugin := range m.plugins {
		items := plugin.GetAdminMenuItems()
		allItems = append(allItems, items...)
	}

	return allItems
}

func (m *Manager) UnloadPlugin(name string) error {
	plugin, exists := m.plugins[name]
	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Shutdown the plugin
	if err := plugin.Shutdown(); err != nil {
		log.Printf("Error shutting down plugin %s: %v", name, err)
	}

	// Remove from manager
	delete(m.plugins, name)
	delete(m.pluginPaths, name)

	log.Printf("Unloaded plugin: %s", name)
	return nil
}

func (m *Manager) ReloadPlugin(name string) error {
	// Get the plugin path before unloading
	path, exists := m.pluginPaths[name]
	if !exists {
		return fmt.Errorf("plugin %s path not found", name)
	}

	// Unload the plugin
	if err := m.UnloadPlugin(name); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// Load it again
	return m.LoadPlugin(path)
}

func (m *Manager) GetPluginSettings(pluginName string) ([]PluginSetting, error) {
	plugin, exists := m.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return plugin.GetSettings(), nil
}

func (m *Manager) ShutdownAll() {
	for name, plugin := range m.plugins {
		if err := plugin.Shutdown(); err != nil {
			log.Printf("Error shutting down plugin %s: %v", name, err)
		}
	}
}
