package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"plugin"
	"runtime"
	"strings"
	"time"
)

type Loader struct {
	loadedPlugins map[string]*plugin.Plugin
	pluginDir     string
	buildDir      string
	extractor     *Extractor
	compiler      *Compiler
}

func NewLoader(pluginDir string) *Loader {
	buildDir := filepath.Join(pluginDir, ".build")

	return &Loader{
		loadedPlugins: make(map[string]*plugin.Plugin),
		pluginDir:     pluginDir,
		buildDir:      buildDir,
		extractor:     NewExtractor(pluginDir),
		compiler:      NewCompiler(buildDir),
	}
}

// InstallFromZip installs a plugin from a zip file
func (l *Loader) InstallFromZip(zipPath, pluginName string) error {
	// Extract the zip file
	pluginDir, err := l.extractor.ExtractZipPlugin(zipPath, pluginName)
	if err != nil {
		return fmt.Errorf("failed to extract plugin: %w", err)
	}

	// Validate plugin structure
	if err := l.extractor.ValidatePluginStructure(pluginDir); err != nil {
		// Clean up on failure
		os.RemoveAll(pluginDir)
		return fmt.Errorf("invalid plugin structure: %w", err)
	}

	// Compile the plugin
	soPath, recompiled, err := l.compiler.CompileWithCache(pluginDir, pluginName)
	if err != nil {
		// Clean up on failure
		os.RemoveAll(pluginDir)
		return fmt.Errorf("failed to compile plugin: %w", err)
	}

	// Validate compilation
	if err := l.compiler.ValidateCompilation(soPath); err != nil {
		os.RemoveAll(pluginDir)
		os.Remove(soPath)
		return fmt.Errorf("plugin compilation validation failed: %w", err)
	}

	// Load the compiled plugin
	if err := l.LoadPluginFromFile(soPath); err != nil {
		os.RemoveAll(pluginDir)
		os.Remove(soPath)
		return fmt.Errorf("failed to load compiled plugin: %w", err)
	}

	// Clean up the original zip file
	os.Remove(zipPath)

	if recompiled {
		fmt.Printf("Plugin %s installed and compiled successfully\n", pluginName)
	} else {
		fmt.Printf("Plugin %s installed (using cached build)\n", pluginName)
	}

	return nil
}

// LoadPluginFromFile loads a plugin from a .so file (maintains backward compatibility)
func (l *Loader) LoadPluginFromFile(path string) error {
	// Check if we're on a supported platform
	if !l.isPluginSupported() {
		return fmt.Errorf("plugins are not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("plugin file not found: %s", path)
	}

	// Load the plugin
	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open plugin %s: %w", path, err)
	}

	// Store loaded plugin for potential cleanup
	pluginName := l.getPluginNameFromPath(path)
	l.loadedPlugins[pluginName] = p

	// Look for the NewPlugin symbol
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return fmt.Errorf("plugin %s must export 'NewPlugin' function: %w", path, err)
	}

	// Verify the symbol is the correct type
	newPluginFunc, ok := symbol.(func() Plugin)
	if !ok {
		return fmt.Errorf("NewPlugin in %s must be a function that returns Plugin interface", path)
	}

	// Create plugin instance to validate it
	pluginInstance := newPluginFunc()
	if pluginInstance == nil {
		return fmt.Errorf("NewPlugin function in %s returned nil", path)
	}

	return nil
}

// LoadPluginFromDirectory loads a plugin from source directory
func (l *Loader) LoadPluginFromDirectory(pluginName string) (Plugin, error) {
	pluginDir := filepath.Join(l.pluginDir, pluginName)

	// Check if plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin directory not found: %s", pluginDir)
	}

	// Compile the plugin
	soPath, _, err := l.compiler.CompileWithCache(pluginDir, pluginName)
	if err != nil {
		return nil, fmt.Errorf("failed to compile plugin: %w", err)
	}

	// Load the compiled plugin
	return l.loadCompiledPlugin(soPath)
}

// LoadCompiledPlugin loads a plugin from a compiled .so file and returns the instance
func (l *Loader) loadCompiledPlugin(soPath string) (Plugin, error) {
	// Load the plugin
	p, err := plugin.Open(soPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin %s: %w", soPath, err)
	}

	// Store loaded plugin
	pluginName := l.getPluginNameFromPath(soPath)
	l.loadedPlugins[pluginName] = p

	// Look for the NewPlugin symbol
	symbol, err := p.Lookup("NewPlugin")
	if err != nil {
		return nil, fmt.Errorf("plugin %s must export 'NewPlugin' function: %w", soPath, err)
	}

	// Verify the symbol is the correct type
	newPluginFunc, ok := symbol.(func() Plugin)
	if !ok {
		return nil, fmt.Errorf("NewPlugin in %s must be a function that returns Plugin interface", soPath)
	}

	// Create plugin instance
	pluginInstance := newPluginFunc()
	if pluginInstance == nil {
		return nil, fmt.Errorf("NewPlugin function in %s returned nil", soPath)
	}

	return pluginInstance, nil
}

// LoadAllPlugins discovers and loads all plugins
func (l *Loader) LoadAllPlugins() (map[string]Plugin, error) {
	plugins := make(map[string]Plugin)
	var loadErrors []string

	// Create plugin directory if it doesn't exist
	if err := os.MkdirAll(l.pluginDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Find all plugin directories
	entries, err := os.ReadDir(l.pluginDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue // Skip files and hidden directories
		}

		pluginName := entry.Name()
		pluginInstance, err := l.LoadPluginFromDirectory(pluginName)
		if err != nil {
			loadErrors = append(loadErrors, fmt.Sprintf("%s: %v", pluginName, err))
			continue
		}

		info := pluginInstance.GetInfo()
		plugins[info.Name] = pluginInstance
	}

	// Also load any standalone .so files for backward compatibility
	soFiles, err := filepath.Glob(filepath.Join(l.pluginDir, "*.so"))
	if err == nil {
		for _, soPath := range soFiles {
			pluginName := l.getPluginNameFromPath(soPath)

			// Skip if we already loaded this plugin from directory
			if _, exists := plugins[pluginName]; exists {
				continue
			}

			pluginInstance, err := l.loadCompiledPlugin(soPath)
			if err != nil {
				loadErrors = append(loadErrors, fmt.Sprintf("%s: %v", pluginName, err))
				continue
			}

			info := pluginInstance.GetInfo()
			plugins[info.Name] = pluginInstance
		}
	}

	// Return errors if any plugins failed to load
	if len(loadErrors) > 0 {
		return plugins, fmt.Errorf("failed to load some plugins:\n%s", strings.Join(loadErrors, "\n"))
	}

	return plugins, nil
}

// ValidateZipPlugin validates a zip file before installation
func (l *Loader) ValidateZipPlugin(zipPath string) (*PluginValidationResult, error) {
	result := &PluginValidationResult{
		IsValid:  true,
		Errors:   []string{},
		Warnings: []string{},
	}

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(zipPath), ".zip") {
		result.IsValid = false
		result.Errors = append(result.Errors, "File must be a .zip archive")
		return result, nil
	}

	// Check file size (100MB limit)
	info, err := os.Stat(zipPath)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Cannot access file: %v", err))
		return result, nil
	}

	if info.Size() > 100*1024*1024 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Plugin file too large (max 100MB)")
		return result, nil
	}

	if info.Size() == 0 {
		result.IsValid = false
		result.Errors = append(result.Errors, "Plugin file is empty")
		return result, nil
	}

	// TODO: Add more validation like checking zip contents without extracting

	return result, nil
}

// GetPluginInfo gets plugin information without loading it
func (l *Loader) GetPluginInfo(pluginName string) (*PluginInfo, error) {
	pluginDir := filepath.Join(l.pluginDir, pluginName)

	// Check if plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	// Try to read plugin.json first
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var manifest PluginManifest
		if err := json.Unmarshal(data, &manifest); err == nil {
			return &PluginInfo{
				Name:        manifest.Name,
				Version:     manifest.Version,
				Description: manifest.Description,
				Author:      manifest.Author,
				Website:     manifest.Website,
			}, nil
		}
	}

	// Fallback: compile and load plugin to get info
	pluginInstance, err := l.LoadPluginFromDirectory(pluginName)
	if err != nil {
		return nil, err
	}

	info := pluginInstance.GetInfo()
	return &info, nil
}

// UninstallPlugin removes a plugin completely
func (l *Loader) UninstallPlugin(pluginName string) error {
	pluginDir := filepath.Join(l.pluginDir, pluginName)
	soPath := filepath.Join(l.buildDir, pluginName+".so")

	// Remove plugin directory
	if err := os.RemoveAll(pluginDir); err != nil {
		return fmt.Errorf("failed to remove plugin directory: %w", err)
	}

	// Remove compiled .so file
	os.Remove(soPath)

	// Remove from loaded plugins
	delete(l.loadedPlugins, pluginName)

	return nil
}

// RecompilePlugin forces recompilation of a plugin
func (l *Loader) RecompilePlugin(pluginName string) error {
	pluginDir := filepath.Join(l.pluginDir, pluginName)
	soPath := filepath.Join(l.buildDir, pluginName+".so")

	// Remove existing compiled file to force recompilation
	os.Remove(soPath)

	// Recompile
	_, err := l.compiler.CompilePlugin(pluginDir, pluginName)
	return err
}

// Utility functions

func (l *Loader) isPluginSupported() bool {
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

// GetSupportedPlatforms returns list of supported platforms
func (l *Loader) GetSupportedPlatforms() []string {
	return []string{"linux", "darwin", "freebsd"}
}

// GetCurrentPlatform returns current platform info
func (l *Loader) GetCurrentPlatform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}

// IsPlatformSupported checks if current platform supports plugins
func (l *Loader) IsPlatformSupported() bool {
	return l.isPluginSupported()
}

// CleanupBuildCache removes old compiled files
func (l *Loader) CleanupBuildCache(maxAge time.Duration) error {
	return l.compiler.CleanupOldBuilds(maxAge)
}

// GetCompilerInfo returns information about the Go compiler
func (l *Loader) GetCompilerInfo() (*CompilerInfo, error) {
	return l.compiler.GetCompilerInfo()
}

// Supporting types

type PluginValidationResult struct {
	IsValid  bool     `json:"is_valid"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

type CompatibilityInfo struct {
	Path            string      `json:"path"`
	CurrentPlatform string      `json:"current_platform"`
	IsSupported     bool        `json:"is_supported"`
	IsCompatible    bool        `json:"is_compatible"`
	Issues          []string    `json:"issues,omitempty"`
	PluginInfo      *PluginInfo `json:"plugin_info,omitempty"`
}
