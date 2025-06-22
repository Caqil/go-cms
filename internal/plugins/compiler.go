package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Compiler struct {
	buildDir string
	goPath   string
}

func NewCompiler(buildDir string) *Compiler {
	return &Compiler{
		buildDir: buildDir,
		goPath:   "go", // Can be overridden if Go is not in PATH
	}
}

// CompilePlugin compiles a plugin directory into a .so file
func (c *Compiler) CompilePlugin(pluginDir, pluginName string) (string, error) {
	// Ensure build directory exists
	if err := os.MkdirAll(c.buildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}

	// Output file path
	outputFile := filepath.Join(c.buildDir, pluginName+".so")

	// Remove existing build
	os.Remove(outputFile)

	// Check if we need to run go mod tidy
	if err := c.ensureGoMod(pluginDir, pluginName); err != nil {
		return "", fmt.Errorf("failed to setup go module: %w", err)
	}

	// Build the plugin
	cmd := exec.Command(c.goPath, "build", "-buildmode=plugin", "-o", outputFile, ".")
	cmd.Dir = pluginDir

	// Set environment variables
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=1", // Required for plugin mode
		"GO111MODULE=on",
	)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("compilation failed: %s\nOutput: %s", err, string(output))
	}

	// Verify the output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return "", fmt.Errorf("compilation succeeded but output file not found: %s", outputFile)
	}

	return outputFile, nil
}

// EnsureGoMod ensures the plugin has a proper go.mod file
func (c *Compiler) ensureGoMod(pluginDir, pluginName string) error {
	goModPath := filepath.Join(pluginDir, "go.mod")

	// Check if go.mod exists
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		// Create go.mod
		if err := c.createGoMod(pluginDir, pluginName); err != nil {
			return err
		}
	}

	// Run go mod tidy to ensure dependencies are resolved
	cmd := exec.Command(c.goPath, "mod", "tidy")
	cmd.Dir = pluginDir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s\nOutput: %s", err, string(output))
	}

	return nil
}

// CreateGoMod creates a go.mod file for the plugin
func (c *Compiler) createGoMod(pluginDir, pluginName string) error {
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	go.mongodb.org/mongo-driver v1.12.1
)
`, pluginName)

	goModPath := filepath.Join(pluginDir, "go.mod")
	return os.WriteFile(goModPath, []byte(goModContent), 0644)
}

// CompileWithCache compiles plugin only if source is newer than built .so file
func (c *Compiler) CompileWithCache(pluginDir, pluginName string) (string, bool, error) {
	outputFile := filepath.Join(c.buildDir, pluginName+".so")

	// Check if .so file exists and get its modification time
	soInfo, err := os.Stat(outputFile)
	if os.IsNotExist(err) {
		// File doesn't exist, need to compile
		compiled, err := c.CompilePlugin(pluginDir, pluginName)
		return compiled, true, err
	}

	// Check if any source files are newer than the .so file
	needsRecompile, err := c.needsRecompilation(pluginDir, soInfo.ModTime())
	if err != nil {
		return "", false, err
	}

	if needsRecompile {
		compiled, err := c.CompilePlugin(pluginDir, pluginName)
		return compiled, true, err
	}

	// Return existing file
	return outputFile, false, nil
}

// NeedsRecompilation checks if any source files are newer than the target
func (c *Compiler) needsRecompilation(pluginDir string, targetTime time.Time) (bool, error) {
	needsRecompile := false

	err := filepath.Walk(pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check .go files and go.mod
		if strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "go.mod") || strings.HasSuffix(path, "go.sum") {
			if info.ModTime().After(targetTime) {
				needsRecompile = true
				return filepath.SkipDir // Found a newer file, can stop checking
			}
		}

		return nil
	})

	return needsRecompile, err
}

// ValidateCompilation validates that the compiled plugin can be loaded
func (c *Compiler) ValidateCompilation(soPath string) error {
	// Try to open the plugin to verify it's valid
	// This doesn't actually load it, just checks if it's a valid plugin
	cmd := exec.Command(c.goPath, "tool", "objdump", "-t", soPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("plugin validation failed: %s", string(output))
	}

	// Check for required symbols
	outputStr := string(output)
	if !strings.Contains(outputStr, "NewPlugin") {
		return fmt.Errorf("plugin does not export required NewPlugin function")
	}

	return nil
}

// CleanupOldBuilds removes old .so files to save space
func (c *Compiler) CleanupOldBuilds(maxAge time.Duration) error {
	entries, err := os.ReadDir(c.buildDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".so") {
			filePath := filepath.Join(c.buildDir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				os.Remove(filePath)
			}
		}
	}

	return nil
}

// GetCompilerInfo returns information about the Go compiler
func (c *Compiler) GetCompilerInfo() (*CompilerInfo, error) {
	// Get Go version
	cmd := exec.Command(c.goPath, "version")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Go version: %w", err)
	}

	// Get Go environment
	cmd = exec.Command(c.goPath, "env", "GOOS", "GOARCH", "CGO_ENABLED")
	envOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get Go environment: %w", err)
	}

	envLines := strings.Split(strings.TrimSpace(string(envOutput)), "\n")

	return &CompilerInfo{
		Version:    strings.TrimSpace(string(output)),
		GOOS:       envLines[0],
		GOARCH:     envLines[1],
		CGOEnabled: envLines[2] == "1",
		BuildDir:   c.buildDir,
	}, nil
}

type CompilerInfo struct {
	Version    string `json:"version"`
	GOOS       string `json:"goos"`
	GOARCH     string `json:"goarch"`
	CGOEnabled bool   `json:"cgo_enabled"`
	BuildDir   string `json:"build_dir"`
}
