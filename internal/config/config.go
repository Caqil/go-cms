package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         string
	MongoURL     string
	DatabaseName string
	JWTSecret    string
	Environment  string
	PluginPath   string
	ThemePath    string
	AdminPath    string
}

func Load() (*Config, error) {
	return &Config{
		Port:         getEnv("PORT", "8080"),
		MongoURL:     getEnv("MONGO_URL", "mongodb://localhost:27017"),
		DatabaseName: getEnv("DATABASE_NAME", "cms_db"),
		JWTSecret:    getEnv("JWT_SECRET", "your-secret-key"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		PluginPath:   getEnv("PLUGIN_PATH", "./plugins"),
		ThemePath:    getEnv("THEME_PATH", "./themes"),
		AdminPath:    getEnv("ADMIN_PATH", "./web/admin"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
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
