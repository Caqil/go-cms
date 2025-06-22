package themes

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"go-cms/internal/database"
	"go-cms/internal/database/models"

	"go.mongodb.org/mongo-driver/bson"
)

type Manager struct {
	themes    map[string]*Theme
	themePath string
	active    string
	db        *database.DB
}

type Theme struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Description     string            `json:"description"`
	Author          string            `json:"author"`
	AuthorURI       string            `json:"author_uri,omitempty"`
	Screenshot      string            `json:"screenshot,omitempty"`
	Tags            []string          `json:"tags,omitempty"`
	MinVersion      string            `json:"min_version,omitempty"`
	RequiredPlugins []string          `json:"required_plugins,omitempty"`
	Assets          map[string]string `json:"assets"`
	Templates       []Template        `json:"templates,omitempty"`
	Customization   Customization     `json:"customization,omitempty"`
	Path            string            `json:"-"`
	IsActive        bool              `json:"is_active"`
	InstalledAt     time.Time         `json:"installed_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type Template struct {
	Name        string `json:"name"`
	File        string `json:"file"`
	Description string `json:"description"`
	Type        string `json:"type"` // page, post, archive, etc.
}

type Customization struct {
	Colors    map[string]string      `json:"colors,omitempty"`
	Fonts     map[string]string      `json:"fonts,omitempty"`
	Layout    map[string]interface{} `json:"layout,omitempty"`
	CustomCSS string                 `json:"custom_css,omitempty"`
	CustomJS  string                 `json:"custom_js,omitempty"`
}

func NewManager(themePath string, db *database.DB) *Manager {
	return &Manager{
		themes:    make(map[string]*Theme),
		themePath: themePath,
		active:    "default",
		db:        db,
	}
}

func (m *Manager) LoadThemes() error {
	// Read theme directories
	dirs, err := ioutil.ReadDir(m.themePath)
	if err != nil {
		return fmt.Errorf("failed to read theme directory: %w", err)
	}

	for _, dir := range dirs {
		if dir.IsDir() {
			themePath := filepath.Join(m.themePath, dir.Name())
			if err := m.loadTheme(themePath); err != nil {
				fmt.Printf("Failed to load theme %s: %v\n", dir.Name(), err)
			}
		}
	}

	// Load active theme from database
	if err := m.loadActiveThemeFromDB(); err != nil {
		fmt.Printf("Warning: Failed to load active theme from database: %v\n", err)
	}

	return nil
}

func (m *Manager) loadTheme(path string) error {
	metadataPath := filepath.Join(path, "metadata.json")
	data, err := ioutil.ReadFile(metadataPath)
	if err != nil {
		return fmt.Errorf("failed to read metadata.json: %w", err)
	}

	var theme Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return fmt.Errorf("failed to parse metadata.json: %w", err)
	}

	theme.Path = path
	theme.InstalledAt = time.Now()
	theme.UpdatedAt = time.Now()

	// Load theme settings from database if exists
	if m.db != nil {
		m.loadThemeFromDB(&theme)
	}

	m.themes[theme.Name] = &theme
	return nil
}

func (m *Manager) loadThemeFromDB(theme *Theme) error {
	collection := m.db.Collection("themes")
	var dbTheme models.ThemeMetadata
	err := collection.FindOne(context.Background(), bson.M{"name": theme.Name}).Decode(&dbTheme)
	if err != nil {
		return err // Theme not in database yet
	}

	// Update theme with database information
	theme.IsActive = dbTheme.IsActive
	theme.InstalledAt = dbTheme.InstalledAt
	theme.UpdatedAt = dbTheme.UpdatedAt
	theme.Customization = Customization{
		Colors:    dbTheme.Customization.Colors,
		Fonts:     dbTheme.Customization.Fonts,
		Layout:    dbTheme.Customization.Layout,
		CustomCSS: dbTheme.Customization.CustomCSS,
		CustomJS:  dbTheme.Customization.CustomJS,
	}

	return nil
}

func (m *Manager) loadActiveThemeFromDB() error {
	if m.db == nil {
		return nil
	}

	collection := m.db.Collection("themes")
	var activeTheme models.ThemeMetadata
	err := collection.FindOne(context.Background(), bson.M{"is_active": true}).Decode(&activeTheme)
	if err != nil {
		return err
	}

	m.active = activeTheme.Name
	return nil
}

func (m *Manager) GetTheme(name string) (*Theme, bool) {
	theme, exists := m.themes[name]
	return theme, exists
}

func (m *Manager) GetAllThemes() map[string]*Theme {
	return m.themes
}

func (m *Manager) SetActiveTheme(name string) error {
	theme, exists := m.themes[name]
	if !exists {
		return fmt.Errorf("theme %s not found", name)
	}

	// Check if theme requirements are met
	if err := m.validateThemeRequirements(theme); err != nil {
		return fmt.Errorf("theme requirements not met: %w", err)
	}

	// Update database
	if m.db != nil {
		if err := m.setActiveThemeInDB(name); err != nil {
			return fmt.Errorf("failed to update database: %w", err)
		}
	}

	m.active = name
	return nil
}

func (m *Manager) validateThemeRequirements(theme *Theme) error {
	// Check minimum CMS version
	if theme.MinVersion != "" {
		// In a real implementation, you'd compare versions
		// For now, we'll assume it's compatible
	}

	// Check required plugins
	if len(theme.RequiredPlugins) > 0 {
		// In a real implementation, you'd check if required plugins are installed and active
		// For now, we'll skip this check
	}

	return nil
}

func (m *Manager) setActiveThemeInDB(name string) error {
	collection := m.db.Collection("themes")
	ctx := context.Background()

	// Deactivate all themes
	_, err := collection.UpdateMany(ctx, bson.M{}, bson.M{
		"$set": bson.M{"is_active": false, "updated_at": time.Now()},
	})
	if err != nil {
		return err
	}

	// Activate the selected theme
	filter := bson.M{"name": name}
	update := bson.M{
		"$set": bson.M{
			"is_active":  true,
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// If theme doesn't exist in database, create it
	if result.MatchedCount == 0 {
		theme, exists := m.themes[name]
		if !exists {
			return fmt.Errorf("theme not found in memory")
		}

		themeMetadata := models.ThemeMetadata{
			Name:        theme.Name,
			Version:     theme.Version,
			Description: theme.Description,
			Author:      theme.Author,
			Path:        theme.Path,
			IsActive:    true,
			Customization: models.ThemeCustomization{
				Colors:    theme.Customization.Colors,
				Fonts:     theme.Customization.Fonts,
				Layout:    theme.Customization.Layout,
				CustomCSS: theme.Customization.CustomCSS,
				CustomJS:  theme.Customization.CustomJS,
			},
			InstalledAt: time.Now(),
			UpdatedAt:   time.Now(),
		}

		_, err = collection.InsertOne(ctx, themeMetadata)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) GetActiveTheme() *Theme {
	if theme, exists := m.themes[m.active]; exists {
		return theme
	}
	return nil
}

func (m *Manager) GetThemeCustomization(name string) (Customization, error) {
	theme, exists := m.themes[name]
	if !exists {
		return Customization{}, fmt.Errorf("theme not found")
	}

	return theme.Customization, nil
}

func (m *Manager) UpdateThemeCustomization(name string, customization Customization) error {
	theme, exists := m.themes[name]
	if !exists {
		return fmt.Errorf("theme not found")
	}

	// Update in memory
	theme.Customization = customization
	theme.UpdatedAt = time.Now()

	// Update in database
	if m.db != nil {
		return m.updateThemeCustomizationInDB(name, customization)
	}

	return nil
}

func (m *Manager) updateThemeCustomizationInDB(name string, customization Customization) error {
	collection := m.db.Collection("themes")

	filter := bson.M{"name": name}
	update := bson.M{
		"$set": bson.M{
			"customization.colors":     customization.Colors,
			"customization.fonts":      customization.Fonts,
			"customization.layout":     customization.Layout,
			"customization.custom_css": customization.CustomCSS,
			"customization.custom_js":  customization.CustomJS,
			"updated_at":               time.Now(),
		},
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (m *Manager) InstallTheme(themePath string) error {
	// Extract theme to themes directory
	themeName := filepath.Base(themePath)
	themeDir := filepath.Join(m.themePath, themeName)

	// Create theme directory
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		return fmt.Errorf("failed to create theme directory: %w", err)
	}

	// In a real implementation, you'd extract the theme package here
	// For now, we'll assume the theme is already in the correct location

	// Load the theme
	return m.loadTheme(themeDir)
}

func (m *Manager) UninstallTheme(name string) error {
	theme, exists := m.themes[name]
	if !exists {
		return fmt.Errorf("theme not found")
	}

	// Don't allow uninstalling active theme
	if theme.IsActive {
		return fmt.Errorf("cannot uninstall active theme")
	}

	// Remove from database
	if m.db != nil {
		collection := m.db.Collection("themes")
		_, err := collection.DeleteOne(context.Background(), bson.M{"name": name})
		if err != nil {
			return fmt.Errorf("failed to remove theme from database: %w", err)
		}
	}

	// Remove theme directory
	if err := os.RemoveAll(theme.Path); err != nil {
		return fmt.Errorf("failed to remove theme directory: %w", err)
	}

	// Remove from memory
	delete(m.themes, name)

	return nil
}

func (m *Manager) GetThemeAssets(name string) (map[string]string, error) {
	theme, exists := m.themes[name]
	if !exists {
		return nil, fmt.Errorf("theme not found")
	}

	assets := make(map[string]string)

	// Build full paths for assets
	for assetType, assetPath := range theme.Assets {
		fullPath := filepath.Join(theme.Path, assetPath)
		assets[assetType] = fullPath
	}

	return assets, nil
}
