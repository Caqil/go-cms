package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

const pluginTemplate = `package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// {{.PluginStruct}} implements the Plugin interface
type {{.PluginStruct}} struct {
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
	Name        string ` + "`json:\"name\"`" + `
	Version     string ` + "`json:\"version\"`" + `
	Description string ` + "`json:\"description\"`" + `
	Author      string ` + "`json:\"author\"`" + `
	Website     string ` + "`json:\"website,omitempty\"`" + `
}

type AdminMenuItem struct {
	ID       string          ` + "`json:\"id\"`" + `
	Title    string          ` + "`json:\"title\"`" + `
	Icon     string          ` + "`json:\"icon,omitempty\"`" + `
	URL      string          ` + "`json:\"url\"`" + `
	Parent   string          ` + "`json:\"parent,omitempty\"`" + `
	Order    int             ` + "`json:\"order\"`" + `
	Children []AdminMenuItem ` + "`json:\"children,omitempty\"`" + `
}

type PluginSetting struct {
	Key         string      ` + "`json:\"key\"`" + `
	Label       string      ` + "`json:\"label\"`" + `
	Type        string      ` + "`json:\"type\"`" + `
	Value       interface{} ` + "`json:\"value\"`" + `
	Description string      ` + "`json:\"description,omitempty\"`" + `
	Options     []string    ` + "`json:\"options,omitempty\"`" + `
	Required    bool        ` + "`json:\"required\"`" + `
}

// NewPlugin is the entry point that will be called by the plugin manager
func NewPlugin() Plugin {
	return &{{.PluginStruct}}{
		settings: make(map[string]interface{}),
	}
}

func (p *{{.PluginStruct}}) GetInfo() PluginInfo {
	return PluginInfo{
		Name:        "{{.PluginName}}",
		Version:     "{{.Version}}",
		Description: "{{.Description}}",
		Author:      "{{.Author}}",
		Website:     "{{.Website}}",
	}
}

func (p *{{.PluginStruct}}) Initialize(deps *PluginDependencies) error {
	p.deps = deps
	p.setDefaultSettings()
	
	// Perform any initialization here
	// You could set up database connections, load configuration, etc.
	return nil
}

func (p *{{.PluginStruct}}) setDefaultSettings() {
	p.settings = map[string]interface{}{
		"enabled":      true,
		"auto_update":  false,
		"cache_ttl":    3600,
		"debug_mode":   false,
	}
}

func (p *{{.PluginStruct}}) RegisterRoutes(router *gin.RouterGroup) {
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

func (p *{{.PluginStruct}}) GetAdminMenuItems() []AdminMenuItem {
	return []AdminMenuItem{
		{
			ID:    "{{.PluginName}}-menu",
			Title: "{{.MenuTitle}}",
			Icon:  "{{.MenuIcon}}",
			URL:   "/admin/plugins/{{.PluginName}}",
			Order: {{.MenuOrder}},
			Children: []AdminMenuItem{
				{
					ID:    "{{.PluginName}}-dashboard",
					Title: "Dashboard",
					URL:   "/admin/plugins/{{.PluginName}}/admin/dashboard",
					Order: 1,
				},
				{
					ID:    "{{.PluginName}}-settings",
					Title: "Settings",
					URL:   "/admin/plugins/{{.PluginName}}/admin/settings",
					Order: 2,
				},
			},
		},
	}
}

func (p *{{.PluginStruct}}) GetSettings() []PluginSetting {
	return []PluginSetting{
		{
			Key:         "enabled",
			Label:       "Enable {{.MenuTitle}}",
			Type:        "boolean",
			Value:       p.settings["enabled"],
			Description: "Enable or disable {{.MenuTitle}} functionality",
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

func (p *{{.PluginStruct}}) Shutdown() error {
	// Perform cleanup here
	return nil
}

// HTTP Handlers

func (p *{{.PluginStruct}}) handleIndex(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugin":     p.GetInfo().Name,
		"version":    p.GetInfo().Version,
		"message":    "Welcome to {{.MenuTitle}}!",
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

func (p *{{.PluginStruct}}) handleInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"info":     p.GetInfo(),
		"settings": p.GetSettings(),
	})
}

func (p *{{.PluginStruct}}) handleStatus(c *gin.Context) {
	enabled := p.settings["enabled"].(bool)
	
	c.JSON(http.StatusOK, gin.H{
		"plugin":    p.GetInfo().Name,
		"version":   p.GetInfo().Version,
		"enabled":   enabled,
		"uptime":    time.Now().Format(time.RFC3339),
		"status":    "healthy",
	})
}

func (p *{{.PluginStruct}}) handleAction(c *gin.Context) {
	var requestData struct {
		Action string                 ` + "`json:\"action\"` binding:\"required\"`" + `
		Data   map[string]interface{} ` + "`json:\"data\"`" + `
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

func (p *{{.PluginStruct}}) handleAdminDashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugin":  p.GetInfo().Name,
		"title":   "{{.MenuTitle}} Dashboard",
		"stats": map[string]interface{}{
			"enabled":    p.settings["enabled"],
			"version":    p.GetInfo().Version,
			"cache_ttl":  p.settings["cache_ttl"],
			"debug_mode": p.settings["debug_mode"],
		},
	})
}

func (p *{{.PluginStruct}}) handleAdminSettings(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"plugin":   p.GetInfo().Name,
		"settings": p.settings,
	})
}

func (p *{{.PluginStruct}}) handleUpdateSettings(c *gin.Context) {
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
`

const goModTemplate = `module {{.PluginName}}

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	go.mongodb.org/mongo-driver v1.12.1
)
`

const pluginManifestTemplate = `{
  "name": "{{.PluginName}}",
  "version": "{{.Version}}",
  "description": "{{.Description}}",
  "author": "{{.Author}}",
  "website": "{{.Website}}",
  "main": "main.go",
  "dependencies": {
    "go": "1.21",
    "gin": "v1.9.1"
  },
  "scripts": {
    "build": "go build -buildmode=plugin -o {{.PluginName}}.so .",
    "test": "go test ./...",
    "clean": "rm -f {{.PluginName}}.so"
  },
  "keywords": ["cms", "plugin", "{{.PluginName}}"],
  "license": "MIT"
}
`

const makefileTemplate = `.PHONY: build clean test zip

PLUGIN_NAME={{.PluginName}}
OUTPUT_FILE=$(PLUGIN_NAME).so

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@go build -buildmode=plugin -o $(OUTPUT_FILE) .
	@echo "✓ Plugin built successfully: $(OUTPUT_FILE)"

clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(OUTPUT_FILE)
	@rm -f $(PLUGIN_NAME).zip
	@echo "✓ Clean completed"

test:
	@echo "Running tests..."
	@go test ./...

zip: clean
	@echo "Creating plugin zip package..."
	@zip -r $(PLUGIN_NAME).zip . -x "*.so" "*.zip" ".git/*" ".DS_Store"
	@echo "✓ Plugin package created: $(PLUGIN_NAME).zip"

install: zip
	@echo "Installing plugin..."
	@curl -X POST \
		-F "plugin=@$(PLUGIN_NAME).zip" \
		http://localhost:8080/api/v1/admin/plugins/upload \
		-H "Authorization: Bearer YOUR_TOKEN_HERE"

dev: build
	@echo "Development build completed"

.DEFAULT_GOAL := build
`

const readmeTemplate = `# {{.MenuTitle}} Plugin

{{.Description}}

## Features

- WordPress-like plugin architecture
- Hot-reloadable without server restart
- Dynamic route registration
- Admin interface integration
- Configurable settings

## Installation

1. Create a zip file with the plugin contents:
   ` + "```bash" + `
   make zip
   ` + "```" + `

2. Upload via admin interface or API:
   ` + "```bash" + `
   curl -X POST \
     -F "plugin=@{{.PluginName}}.zip" \
     http://localhost:8080/api/v1/admin/plugins/upload \
     -H "Authorization: Bearer YOUR_TOKEN"
   ` + "```" + `

## API Endpoints

After installation, the plugin exposes these endpoints:

- ` + "`GET /api/v1/plugins/{{.PluginName}}/`" + ` - Plugin index
- ` + "`GET /api/v1/plugins/{{.PluginName}}/info`" + ` - Plugin information
- ` + "`GET /api/v1/plugins/{{.PluginName}}/status`" + ` - Plugin status
- ` + "`POST /api/v1/plugins/{{.PluginName}}/action`" + ` - Execute actions

## Admin Interface

The plugin adds menu items to the admin dashboard:

- **{{.MenuTitle}}** > Dashboard
- **{{.MenuTitle}}** > Settings

## Development

1. Build the plugin:
   ` + "```bash" + `
   make build
   ` + "```" + `

2. Test the plugin:
   ` + "```bash" + `
   make test
   ` + "```" + `

3. Create distributable package:
   ` + "```bash" + `
   make zip
   ` + "```" + `

## Configuration

The plugin supports these settings:

- **enabled**: Enable/disable plugin functionality
- **auto_update**: Automatic updates when available
- **cache_ttl**: Cache time-to-live in seconds
- **debug_mode**: Enable debug logging

## Author

{{.Author}}

## License

MIT License
`

type PluginConfig struct {
	PluginName   string
	PluginStruct string
	Version      string
	Description  string
	Author       string
	Website      string
	MenuTitle    string
	MenuIcon     string
	MenuOrder    int
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: plugin-builder <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  create <name>     Create a new plugin")
		fmt.Println("  build <name>      Build an existing plugin")
		fmt.Println("  build-all         Build all plugins")
		fmt.Println("  zip <name>        Create zip package for plugin")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Usage: plugin-builder create <plugin-name>")
		}
		createPlugin(os.Args[2])
	case "build":
		if len(os.Args) < 3 {
			log.Fatal("Usage: plugin-builder build <plugin-name>")
		}
		buildPlugin(os.Args[2])
	case "build-all":
		buildAllPlugins()
	case "zip":
		if len(os.Args) < 3 {
			log.Fatal("Usage: plugin-builder zip <plugin-name>")
		}
		createZipPackage(os.Args[2])
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func createPlugin(name string) {
	// Validate plugin name
	if !isValidPluginName(name) {
		log.Fatal("Invalid plugin name. Use lowercase letters, numbers, and hyphens only.")
	}

	// Create plugin directory
	pluginDir := filepath.Join("plugins", name)
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		log.Fatalf("Failed to create plugin directory: %v", err)
	}

	// Generate plugin configuration
	config := PluginConfig{
		PluginName:   name,
		PluginStruct: toPascalCase(name) + "Plugin",
		Version:      "1.0.0",
		Description:  fmt.Sprintf("A %s plugin for the CMS", name),
		Author:       "Plugin Developer",
		Website:      "",
		MenuTitle:    toTitleCase(name),
		MenuIcon:     "puzzle-piece",
		MenuOrder:    50,
	}

	// Create files
	files := map[string]string{
		"main.go":     pluginTemplate,
		"go.mod":      goModTemplate,
		"plugin.json": pluginManifestTemplate,
		"Makefile":    makefileTemplate,
		"README.md":   readmeTemplate,
	}

	for filename, templateStr := range files {
		if err := createFileFromTemplate(
			filepath.Join(pluginDir, filename),
			templateStr,
			config,
		); err != nil {
			log.Fatalf("Failed to create %s: %v", filename, err)
		}
	}

	fmt.Printf("✓ Plugin '%s' created successfully in %s\n", name, pluginDir)
	fmt.Printf("  To build: cd %s && make build\n", pluginDir)
	fmt.Printf("  To create zip: cd %s && make zip\n", pluginDir)
	fmt.Printf("  To install: cd %s && make install\n", pluginDir)
}

func buildPlugin(name string) {
	pluginDir := filepath.Join("plugins", name)

	// Check if plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		log.Fatalf("Plugin directory not found: %s", pluginDir)
	}

	// Check if main.go exists
	mainFile := filepath.Join(pluginDir, "main.go")
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		log.Fatalf("main.go not found in plugin directory: %s", pluginDir)
	}

	fmt.Printf("Building plugin: %s\n", name)

	// Build the plugin
	outputFile := filepath.Join(pluginDir, name+".so")
	cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputFile, ".")
	cmd.Dir = pluginDir

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Failed to build plugin %s:\n%s", name, string(output))
	}

	fmt.Printf("✓ Plugin '%s' built successfully: %s\n", name, outputFile)
}

func buildAllPlugins() {
	pluginsDir := "plugins"

	entries, err := os.ReadDir(pluginsDir)
	if err != nil {
		log.Fatalf("Failed to read plugins directory: %v", err)
	}

	built := 0
	failed := 0

	for _, entry := range entries {
		if entry.IsDir() {
			pluginPath := filepath.Join(pluginsDir, entry.Name())
			mainFile := filepath.Join(pluginPath, "main.go")

			// Check if it's a valid plugin directory
			if _, err := os.Stat(mainFile); err == nil {
				fmt.Printf("Building plugin: %s\n", entry.Name())

				outputFile := filepath.Join(pluginPath, entry.Name()+".so")
				cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputFile, ".")
				cmd.Dir = pluginPath

				if output, err := cmd.CombinedOutput(); err != nil {
					fmt.Printf("✗ Failed to build %s:\n%s\n", entry.Name(), string(output))
					failed++
				} else {
					fmt.Printf("✓ Built %s successfully\n", entry.Name())
					built++
				}
			}
		}
	}

	fmt.Printf("\nBuild summary: %d successful, %d failed\n", built, failed)
}

func createZipPackage(name string) {
	pluginDir := filepath.Join("plugins", name)

	// Check if plugin directory exists
	if _, err := os.Stat(pluginDir); os.IsNotExist(err) {
		log.Fatalf("Plugin directory not found: %s", pluginDir)
	}

	// Create zip package using make command
	cmd := exec.Command("make", "zip")
	cmd.Dir = pluginDir

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Failed to create zip package for %s:\n%s", name, string(output))
	}

	zipFile := filepath.Join(pluginDir, name+".zip")
	fmt.Printf("✓ Plugin package created: %s\n", zipFile)
}

func createFileFromTemplate(path, templateStr string, config PluginConfig) error {
	tmpl, err := template.New("plugin").Parse(templateStr)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, config)
}

func isValidPluginName(name string) bool {
	if len(name) == 0 || len(name) > 50 {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-') {
			return false
		}
	}

	return true
}

func toPascalCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func toTitleCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}
