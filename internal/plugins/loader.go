package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"runtime"
	"strings"
)

type Loader struct {
	loadedPlugins map[string]*plugin.Plugin
	pluginDir     string
}

func NewLoader(pluginDir string) *Loader {
	return &Loader{
		loadedPlugins: make(map[string]*plugin.Plugin),
		pluginDir:     pluginDir,
	}
}

func (l *Loader) LoadPluginFromFile(path string) (Plugin, error) {
	// Check if we're on a supported platform
	if !l.isPluginSupported() {
		return nil, fmt.Errorf("plugins are not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin file not found: %s", path)
	}

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", path, err)
	}

	// Store loaded plugin for potential cleanup
	pluginName := l.getPluginNameFromPath(path)
	l.loadedPlugins[pluginName] = p

	// Look for the NewPlugin symbol
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin %s must export 'NewPlugin' function: %w", path, err)
	}

	// Verify the symbol is the correct type
	newPluginFunc, ok := symbol.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("NewPlugin in %s must be a function that returns Plugin interface", path)
	}

	// Create plugin instance
	pluginInstance := newPluginFunc()
	if pluginInstance == nil {
		return nil, fmt.Errorf("NewPlugin function in %s returned nil", path)
	}

	return pluginInstance, nil
}

func (l *Loader) LoadAllPlugins() (map[string]Plugin, error) {
	plugins := make(map[string]Plugin)

	// Find all .so files in the plugin directory
	pattern := filepath.Join(l.pluginDir, "*.so")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to find plugin files: %w", err)
	}

	var loadErrors []string

	for _, pluginPath := range matches {
		pluginInstance, err := l.LoadPluginFromFile(pluginPath)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("%s: %v", pluginPath, err))
			continue
		}

		info := pluginInstance.GetInfo()
		plugins[info.Name] = pluginInstance
	}

	// Return errors if any plugins failed to load
	if len(loadErrors) > 0 {
		return plugins, fmt.Errorf("failed to load some plugins:\n%s", strings.Join(loadErrors, "\n"))
	}

	return plugins, nil
}

func (l *Loader) ValidatePlugin(path string) error {
	// Check file extension
	if !strings.HasSuffix(path, ".so") {
		return fmt.Errorf("plugin file must have .so extension")
	}

	// Check if file exists and is readable
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access plugin file: %w", err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return fmt.Errorf("plugin path is not a regular file")
	}

	// Check file size (basic sanity check)
	if info.Size() == 0 {
		return fmt.Errorf("plugin file is empty")
	}

	if info.Size() > 100*1024*1024 { // 100MB limit
		return fmt.Errorf("plugin file is too large (max 100MB)")
	}

	return nil
}

func (l *Loader) GetPluginInfo(path string) (*PluginInfo, error) {
	// Validate first
	if err := l.ValidatePlugin(path); err != nil {
		return nil, err
	}

	// Load plugin temporarily to get info
	pluginInstance, err := l.LoadPluginFromFile(path)
	if err != nil {
		return nil, err
	}

	info := pluginInstance.GetInfo()
	return &info, nil
}

func (l *Loader) isPluginSupported() bool {
	// Go plugins only work on:
	// - Linux
	// - macOS (Darwin)
	// - FreeBSD
	supportedOS := map[string]bool{
		"linux":   true,
		"darwin":  true,
		"freebsd": true,
	}

	return supportedOS[runtime.GOOS]
}

func (l *Loader) getPluginNameFromPath(path string) string {
	filename := filepath.Base(path)
	return strings.TrimSuffix(filename, ".so")
}

func (l *Loader) GetSupportedPlatforms() []string {
	return []string{"linux", "darwin", "freebsd"}
}

func (l *Loader) GetCurrentPlatform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

func (l *Loader) IsPlatformSupported() bool {
	return l.isPluginSupported()
}

// PluginCompatibility checks if a plugin is compatible with the current system
func (l *Loader) PluginCompatibility(path string) (*CompatibilityInfo, error) {
	info := &CompatibilityInfo{
		Path:            path,
		CurrentPlatform: l.GetCurrentPlatform(),
		IsSupported:     l.isPluginSupported(),
	}

	if !info.IsSupported {
		info.Issues = append(info.Issues, fmt.Sprintf("Plugins not supported on %s", info.CurrentPlatform))
		return info, nil
	}

	// Validate plugin file
	if err := l.ValidatePlugin(path); err != nil {
		info.Issues = append(info.Issues, err.Error())
	}

	// Try to get plugin info
	pluginInfo, err := l.GetPluginInfo(path)
	if err != nil {
		info.Issues = append(info.Issues, fmt.Sprintf("Cannot load plugin: %v", err))
	} else {
		info.PluginInfo = pluginInfo
	}

	info.IsCompatible = len(info.Issues) == 0

	return info, nil
}

type CompatibilityInfo struct {
	Path            string      `json:"path"`
	CurrentPlatform string      `json:"current_platform"`
	IsSupported     bool        `json:"is_supported"`
	IsCompatible    bool        `json:"is_compatible"`
	Issues          []string    `json:"issues,omitempty"`
	PluginInfo      *PluginInfo `json:"plugin_info,omitempty"`
}
