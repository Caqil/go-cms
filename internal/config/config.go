package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server settings
	Port        string `json:"port"`
	Environment string `json:"environment"`
	Version     string `json:"version"`

	// Database settings
	MongoURI     string `json:"mongo_uri"`
	DatabaseName string `json:"database_name"`

	// Security settings
	JWTSecret string `json:"jwt_secret"`

	// Upload settings
	MaxUploadSize int64         `json:"max_upload_size"`
	UploadTimeout time.Duration `json:"upload_timeout"`
	TempDir       string        `json:"temp_dir"`

	// Logging settings
	LogLevel    string `json:"log_level"`
	EnableDebug bool   `json:"enable_debug"`
	ThemePath   string
	AdminPath   string
	// Plugin settings
	PluginsDir      string `json:"plugins_dir"`
	EnableHotReload bool   `json:"enable_hot_reload"`
}

func Load() (*Config, error) {
	config := &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		Version:         getEnv("VERSION", "1.0.0"),
		MongoURI:        getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DatabaseName:    getEnv("DATABASE_NAME", "cms_db"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		MaxUploadSize:   getEnvInt64("MAX_UPLOAD_SIZE", 100<<20),
		UploadTimeout:   getEnvDuration("UPLOAD_TIMEOUT", 5*time.Minute),
		TempDir:         getEnv("TEMP_DIR", "./temp"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		EnableDebug:     getEnvBool("ENABLE_DEBUG", true),
		PluginsDir:      getEnv("PLUGINS_DIR", "./plugins"),
		EnableHotReload: getEnvBool("ENABLE_HOT_RELOAD", true),
	}

	// Validate critical settings
	if config.JWTSecret == "your-secret-key-change-in-production" && config.Environment == "production" {
		return nil, fmt.Errorf("JWT_SECRET must be set in production environment")
	}

	// Create necessary directories
	createDirIfNotExists(config.TempDir)
	createDirIfNotExists(config.PluginsDir)

	if config.EnableDebug {
		log.Printf("[CONFIG] Loaded configuration: Environment=%s, Port=%s, Debug=%v",
			config.Environment, config.Port, config.EnableDebug)
	}

	return config, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func createDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		log.Printf("[CONFIG] Created directory: %s", dir)
	}
}

// Additional utility functions for debugging

// DebugRequest logs request details for debugging
func DebugRequest(method, path string, headers map[string][]string) {
	if os.Getenv("ENABLE_DEBUG") == "true" {
		log.Printf("[DEBUG] %s %s", method, path)
		for key, values := range headers {
			if key == "Authorization" {
				log.Printf("[DEBUG] Header %s: [REDACTED]", key)
			} else {
				log.Printf("[DEBUG] Header %s: %v", key, values)
			}
		}
	}
}

// CleanupOldTempFiles removes temporary files older than specified duration
func CleanupOldTempFiles(tempDir string, maxAge time.Duration) error {
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-maxAge)
	var removedCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filepath := tempDir + "/" + entry.Name()
			if err := os.Remove(filepath); err == nil {
				removedCount++
			}
		}
	}

	if removedCount > 0 {
		log.Printf("[CLEANUP] Removed %d old temporary files", removedCount)
	}

	return nil
}
