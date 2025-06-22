package plugins

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Manager struct {
	plugins     map[string]Plugin
	pluginPaths map[string]string
	loader      *Loader
	deps        *PluginDependencies
	mu          sync.RWMutex
	router      *gin.RouterGroup // Store router for dynamic route registration
}

func NewManager() *Manager {
	return &Manager{
		plugins:     make(map[string]Plugin),
		pluginPaths: make(map[string]string),
		loader:      NewLoader("./plugins"),
	}
}

// SetDependencies sets the dependencies that will be passed to plugins
func (m *Manager) SetDependencies(deps *PluginDependencies) {
	m.deps = deps
}

// SetRouter stores the router for dynamic route registration
func (m *Manager) SetRouter(router *gin.RouterGroup) {
	m.router = router
}

// InstallPluginFromZip installs a plugin from a zip file
func (m *Manager) InstallPluginFromZip(zipPath, pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate zip file first
	validationResult, err := m.loader.ValidateZipPlugin(zipPath)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !validationResult.IsValid {
		return fmt.Errorf("invalid plugin: %s", strings.Join(validationResult.Errors, ", "))
	}

	// Check if plugin is already loaded
	if _, exists := m.plugins[pluginName]; exists {
		return fmt.Errorf("plugin %s is already installed", pluginName)
	}

	// Install the plugin
	if err := m.loader.InstallFromZip(zipPath, pluginName); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Load the plugin
	pluginInstance, err := m.loader.LoadPluginFromDirectory(pluginName)
	if err != nil {
		return fmt.Errorf("failed to load installed plugin: %w", err)
	}

	// Initialize the plugin
	if m.deps != nil {
		if err := pluginInstance.Initialize(m.deps); err != nil {
			// Cleanup on initialization failure
			m.loader.UninstallPlugin(pluginName)
			return fmt.Errorf("failed to initialize plugin %s: %w", pluginName, err)
		}
	}

	// Store the plugin
	info := pluginInstance.GetInfo()
	m.plugins[info.Name] = pluginInstance
	m.pluginPaths[info.Name] = pluginName

	// Register routes dynamically
	if m.router != nil {
		m.registerPluginRoutes(info.Name, pluginInstance)
	}

	log.Printf("Plugin installed and loaded: %s v%s", info.Name, info.Version)
	return nil
}

// LoadPlugins loads all existing plugins
func (m *Manager) LoadPlugins(pluginDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugins, err := m.loader.LoadAllPlugins()
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	// Initialize all loaded plugins
	for name, plugin := range plugins {
		// Initialize the plugin
		if m.deps != nil {
			if err := plugin.Initialize(m.deps); err != nil {
				log.Printf("Failed to initialize plugin %s: %v", name, err)
				continue
			}
		}

		// Store the plugin
		m.plugins[name] = plugin
		m.pluginPaths[name] = name

		log.Printf("Loaded plugin: %s v%s", plugin.GetInfo().Name, plugin.GetInfo().Version)
	}

	return nil
}

// LoadPlugin loads a single plugin by name
func (m *Manager) LoadPlugin(pluginName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if plugin is already loaded
	if _, exists := m.plugins[pluginName]; exists {
		return fmt.Errorf("plugin %s is already loaded", pluginName)
	}

	// Load the plugin
	pluginInstance, err := m.loader.LoadPluginFromDirectory(pluginName)
	if err != nil {
		return fmt.Errorf("failed to load plugin %s: %w", pluginName, err)
	}

	// Initialize the plugin
	if m.deps != nil {
		if err := pluginInstance.Initialize(m.deps); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", pluginName, err)
		}
	}

	// Store the plugin
	info := pluginInstance.GetInfo()
	m.plugins[info.Name] = pluginInstance
	m.pluginPaths[info.Name] = pluginName

	// Register routes dynamically
	if m.router != nil {
		m.registerPluginRoutes(info.Name, pluginInstance)
	}

	log.Printf("Loaded plugin: %s v%s", info.Name, info.Version)
	return nil
}

// UnloadPlugin unloads a plugin
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

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

	// Note: We can't dynamically remove routes from Gin router
	// This is a limitation of Gin. In a production system, you might
	// need to track routes and rebuild the router or use a more
	// dynamic routing solution.

	log.Printf("Unloaded plugin: %s", name)
	return nil
}

// UninstallPlugin completely removes a plugin
func (m *Manager) UninstallPlugin(name string) error {
	// First unload if loaded
	if _, exists := m.plugins[name]; exists {
		if err := m.UnloadPlugin(name); err != nil {
			return fmt.Errorf("failed to unload plugin: %w", err)
		}
	}

	// Then uninstall from filesystem
	if err := m.loader.UninstallPlugin(name); err != nil {
		return fmt.Errorf("failed to uninstall plugin: %w", err)
	}

	log.Printf("Uninstalled plugin: %s", name)
	return nil
}

// ReloadPlugin reloads a plugin
func (m *Manager) ReloadPlugin(name string) error {
	// Get the plugin path before unloading
	m.mu.RLock()
	pluginPath, exists := m.pluginPaths[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin %s not found", name)
	}

	// Unload the plugin
	if err := m.UnloadPlugin(name); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// Recompile the plugin
	if err := m.loader.RecompilePlugin(pluginPath); err != nil {
		return fmt.Errorf("failed to recompile plugin: %w", err)
	}

	// Load it again
	return m.LoadPlugin(pluginPath)
}

// GetPlugin returns a specific plugin
func (m *Manager) GetPlugin(name string) (Plugin, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[name]
	return plugin, exists
}

// GetAllPlugins returns all loaded plugins
func (m *Manager) GetAllPlugins() map[string]Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	plugins := make(map[string]Plugin)
	for k, v := range m.plugins {
		plugins[k] = v
	}
	return plugins
}

// RegisterRoutes registers all plugin routes
func (m *Manager) RegisterRoutes(router *gin.RouterGroup) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Store router for dynamic registration
	m.router = router

	// Create a plugin group
	router.Group("/plugins")

	for name, plugin := range m.plugins {
		m.registerPluginRoutes(name, plugin)
	}
}

// RegisterPluginRoutes registers routes for a specific plugin
func (m *Manager) registerPluginRoutes(name string, plugin Plugin) {
	if m.router == nil {
		return
	}

	// Create a plugin group
	pluginGroup := m.router.Group("/plugins")

	// Create a sub-group for each plugin
	pluginRouter := pluginGroup.Group("/" + strings.ToLower(name))
	plugin.RegisterRoutes(pluginRouter)
}

// GetAdminMenuItems returns all admin menu items from plugins
func (m *Manager) GetAdminMenuItems() []AdminMenuItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var allItems []AdminMenuItem

	for _, plugin := range m.plugins {
		items := plugin.GetAdminMenuItems()
		allItems = append(allItems, items...)
	}

	return allItems
}

// GetPluginSettings returns settings for a specific plugin
func (m *Manager) GetPluginSettings(pluginName string) ([]PluginSetting, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin %s not found", pluginName)
	}

	return plugin.GetSettings(), nil
}

// GetPluginInfo returns information about an installed plugin (without loading it)
func (m *Manager) GetPluginInfo(pluginName string) (*PluginInfo, error) {
	return m.loader.GetPluginInfo(pluginName)
}

// ListInstalledPlugins returns list of all installed plugins (loaded and unloaded)
func (m *Manager) ListInstalledPlugins() ([]string, error) {
	// This would scan the plugins directory for installed plugins
	// For now, return loaded plugins
	m.mu.RLock()
	defer m.mu.RUnlock()

	var plugins []string
	for name := range m.plugins {
		plugins = append(plugins, name)
	}

	return plugins, nil
}

// ValidatePlugin validates a plugin zip before installation
func (m *Manager) ValidatePlugin(zipPath string) (*PluginValidationResult, error) {
	return m.loader.ValidateZipPlugin(zipPath)
}

// GetSystemInfo returns system information for plugin development
func (m *Manager) GetSystemInfo() (*SystemInfo, error) {
	compilerInfo, err := m.loader.GetCompilerInfo()
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Platform:    m.loader.GetCurrentPlatform(),
		Supported:   m.loader.IsPlatformSupported(),
		Compiler:    *compilerInfo,
		LoadedCount: len(m.plugins),
	}, nil
}

// CleanupCache removes old compiled plugin files
func (m *Manager) CleanupCache(maxAge time.Duration) error {
	return m.loader.CleanupBuildCache(maxAge)
}

// ShutdownAll shuts down all plugins
func (m *Manager) ShutdownAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, plugin := range m.plugins {
		if err := plugin.Shutdown(); err != nil {
			log.Printf("Error shutting down plugin %s: %v", name, err)
		}
	}
}

// HotReload reloads all plugins without restarting the server
func (m *Manager) HotReload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Println("Starting hot reload of all plugins...")

	// Get list of currently loaded plugins
	var pluginNames []string
	for name := range m.plugins {
		pluginNames = append(pluginNames, name)
	}

	// Unload all plugins
	for _, name := range pluginNames {
		if err := m.UnloadPlugin(name); err != nil {
			log.Printf("Error unloading plugin %s during hot reload: %v", name, err)
		}
	}

	// Reload all plugins
	if err := m.LoadPlugins("./plugins"); err != nil {
		log.Printf("Error during hot reload: %v", err)
		return err
	}

	log.Printf("Hot reload completed. Loaded %d plugins.", len(m.plugins))
	return nil
}

// Supporting types

type SystemInfo struct {
	Platform    string       `json:"platform"`
	Supported   bool         `json:"supported"`
	Compiler    CompilerInfo `json:"compiler"`
	LoadedCount int          `json:"loaded_count"`
}
