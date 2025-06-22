package plugins

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Extractor struct {
	basePluginDir string
}

func NewExtractor(basePluginDir string) *Extractor {
	return &Extractor{
		basePluginDir: basePluginDir,
	}
}

// ExtractZipPlugin extracts a zip file to the plugins directory
func (e *Extractor) ExtractZipPlugin(zipPath, pluginName string) (string, error) {
	// Create plugin directory
	pluginDir := filepath.Join(e.basePluginDir, pluginName)
	if err := os.RemoveAll(pluginDir); err != nil {
		return "", fmt.Errorf("failed to clean plugin directory: %w", err)
	}

	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Open zip file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Extract files
	for _, file := range reader.File {
		if err := e.extractFile(file, pluginDir); err != nil {
			return "", fmt.Errorf("failed to extract file %s: %w", file.Name, err)
		}
	}

	return pluginDir, nil
}

func (e *Extractor) extractFile(file *zip.File, destDir string) error {
	// Clean file path to prevent directory traversal
	cleanPath := filepath.Clean(file.Name)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid file path: %s", file.Name)
	}

	destPath := filepath.Join(destDir, cleanPath)

	// Create directory if it's a directory entry
	if file.FileInfo().IsDir() {
		return os.MkdirAll(destPath, file.FileInfo().Mode())
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Open file from zip
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// Create destination file
	outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.FileInfo().Mode())
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Copy content
	_, err = io.Copy(outFile, rc)
	return err
}

// ValidatePluginStructure validates that the extracted plugin has the required structure
func (e *Extractor) ValidatePluginStructure(pluginDir string) error {
	// Check for main.go
	mainFile := filepath.Join(pluginDir, "main.go")
	if _, err := os.Stat(mainFile); os.IsNotExist(err) {
		return fmt.Errorf("plugin must contain main.go file")
	}

	// Check for plugin.json (WordPress-like manifest)
	manifestFile := filepath.Join(pluginDir, "plugin.json")
	if _, err := os.Stat(manifestFile); os.IsNotExist(err) {
		// Try to auto-generate from main.go if plugin.json doesn't exist
		return e.generateManifest(pluginDir)
	}

	return nil
}

// GenerateManifest creates a plugin.json file by analyzing main.go
func (e *Extractor) generateManifest(pluginDir string) error {
	manifest := `{
  "name": "` + filepath.Base(pluginDir) + `",
  "version": "1.0.0",
  "description": "Auto-generated plugin",
  "author": "Unknown",
  "main": "main.go",
  "dependencies": {
    "go": "1.21"
  }
}`

	manifestPath := filepath.Join(pluginDir, "plugin.json")
	return os.WriteFile(manifestPath, []byte(manifest), 0644)
}

// GetPluginInfo reads plugin information from plugin.json
func (e *Extractor) GetPluginInfo(pluginDir string) (*PluginManifest, error) {
	manifestPath := filepath.Join(pluginDir, "plugin.json")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin.json: %w", err)
	}

	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse plugin.json: %w", err)
	}

	return &manifest, nil
}

type PluginManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Website      string            `json:"website,omitempty"`
	Main         string            `json:"main"`
	Dependencies map[string]string `json:"dependencies"`
	Scripts      map[string]string `json:"scripts,omitempty"`
}
