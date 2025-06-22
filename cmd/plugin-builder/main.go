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
    "github.com/gin-gonic/gin"
)

// {{.PluginStruct}} implements the Plugin interface
type {{.PluginStruct}} struct {
    deps *PluginDependencies
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
    return &{{.PluginStruct}}{}
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
    // Perform any initialization here
    return nil
}

func (p *{{.PluginStruct}}) RegisterRoutes(router *gin.RouterGroup) {
    // Register plugin-specific routes
    router.GET("/hello", p.handleHello)
    router.GET("/info", p.handleInfo)
}

func (p *{{.PluginStruct}}) GetAdminMenuItems() []AdminMenuItem {
    return []AdminMenuItem{
        {
            ID:    "{{.PluginName}}-menu",
            Title: "{{.MenuTitle}}",
            Icon:  "{{.MenuIcon}}",
            URL:   "/admin/plugins/{{.PluginName}}",
            Order: {{.MenuOrder}},
        },
    }
}

func (p *{{.PluginStruct}}) GetSettings() []PluginSetting {
    return []PluginSetting{
        {
            Key:         "enabled",
            Label:       "Enable {{.MenuTitle}}",
            Type:        "boolean",
            Value:       true,
            Description: "Enable or disable {{.MenuTitle}} functionality",
            Required:    false,
        },
    }
}

func (p *{{.PluginStruct}}) Shutdown() error {
    // Perform cleanup here
    return nil
}

// HTTP Handlers

func (p *{{.PluginStruct}}) handleHello(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "message": "Hello from {{.MenuTitle}}!",
        "plugin":  p.GetInfo().Name,
        "version": p.GetInfo().Version,
    })
}

func (p *{{.PluginStruct}}) handleInfo(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "info":     p.GetInfo(),
        "settings": p.GetSettings(),
    })
}
`

const goModTemplate = `module {{.PluginName}}

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
)
`

const makefileTemplate = `.PHONY: build clean

PLUGIN_NAME={{.PluginName}}
OUTPUT_FILE=../$(PLUGIN_NAME).so

build:
	@echo "Building $(PLUGIN_NAME) plugin..."
	@go build -buildmode=plugin -o $(OUTPUT_FILE) .
	@echo "✓ Plugin built successfully: $(OUTPUT_FILE)"

clean:
	@rm -f $(OUTPUT_FILE)
	@echo "✓ Cleaned $(PLUGIN_NAME) plugin"

test:
	@go test -v ./...

install: build
	@echo "✓ Plugin installed and ready to use"
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
		fmt.Println("Usage: plugin-builder <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  create <name> - Create a new plugin")
		fmt.Println("  build <name>  - Build an existing plugin")
		fmt.Println("  build-all     - Build all plugins")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		if len(os.Args) < 3 {
			fmt.Println("Usage: plugin-builder create <plugin-name>")
			os.Exit(1)
		}
		createPlugin(os.Args[2])
	case "build":
		if len(os.Args) < 3 {
			fmt.Println("Usage: plugin-builder build <plugin-name>")
			os.Exit(1)
		}
		buildPlugin(os.Args[2])
	case "build-all":
		buildAllPlugins()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
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

	// Create main.go
	if err := createFileFromTemplate(
		filepath.Join(pluginDir, "main.go"),
		pluginTemplate,
		config,
	); err != nil {
		log.Fatalf("Failed to create main.go: %v", err)
	}

	// Create go.mod
	if err := createFileFromTemplate(
		filepath.Join(pluginDir, "go.mod"),
		goModTemplate,
		config,
	); err != nil {
		log.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create Makefile
	if err := createFileFromTemplate(
		filepath.Join(pluginDir, "Makefile"),
		makefileTemplate,
		config,
	); err != nil {
		log.Fatalf("Failed to create Makefile: %v", err)
	}

	fmt.Printf("✓ Plugin '%s' created successfully in %s\n", name, pluginDir)
	fmt.Printf("  To build: cd %s && make build\n", pluginDir)
	fmt.Printf("  To build with plugin-builder: plugin-builder build %s\n", name)
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
	outputFile := filepath.Join("plugins", name+".so")
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

				outputFile := filepath.Join(pluginsDir, entry.Name()+".so")
				cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputFile, ".")
				cmd.Dir = pluginPath

				if output, err := cmd.CombinedOutput(); err != nil {
					fmt.Printf("✗ Failed to build plugin %s:\n%s\n", entry.Name(), string(output))
					failed++
				} else {
					fmt.Printf("✓ Plugin '%s' built successfully\n", entry.Name())
					built++
				}
			}
		}
	}

	fmt.Printf("\nBuild complete: %d successful, %d failed\n", built, failed)
}

func createFileFromTemplate(filename, templateStr string, data interface{}) error {
	tmpl, err := template.New("file").Parse(templateStr)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func isValidPluginName(name string) bool {
	if len(name) == 0 {
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
			parts[i] = strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

func toTitleCase(s string) string {
	parts := strings.Split(s, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(string(part[0])) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}
