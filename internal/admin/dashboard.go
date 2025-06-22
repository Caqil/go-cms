package admin

import (
	"context"
	"fmt"
	"time"

	"go-cms/internal/database"
	"go-cms/internal/plugins"
	"go-cms/internal/themes"

	"go.mongodb.org/mongo-driver/bson"
)

type DashboardData struct {
	Stats          SystemStats    `json:"stats"`
	RecentActivity []Activity     `json:"recent_activity"`
	SystemInfo     SystemInfo     `json:"system_info"`
	PluginStatus   []PluginStatus `json:"plugin_status"`
}

type SystemStats struct {
	TotalUsers    int64  `json:"total_users"`
	ActiveUsers   int64  `json:"active_users"`
	TotalPlugins  int    `json:"total_plugins"`
	ActivePlugins int    `json:"active_plugins"`
	TotalThemes   int    `json:"total_themes"`
	ActiveTheme   string `json:"active_theme"`
	DatabaseSize  int64  `json:"database_size"`
	SystemUptime  string `json:"system_uptime"`
}

type Activity struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
	User        string    `json:"user"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}

type SystemInfo struct {
	Version        string    `json:"version"`
	GoVersion      string    `json:"go_version"`
	StartTime      time.Time `json:"start_time"`
	DatabaseStatus string    `json:"database_status"`
	Environment    string    `json:"environment"`
}

type PluginStatus struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	Status    string `json:"status"`
	LoadTime  string `json:"load_time"`
	LastError string `json:"last_error,omitempty"`
}

type DashboardManager struct {
	db            *database.DB
	pluginManager *plugins.Manager
	themeManager  *themes.Manager
	startTime     time.Time
}

func NewDashboardManager(db *database.DB, pluginManager *plugins.Manager, themeManager *themes.Manager) *DashboardManager {
	return &DashboardManager{
		db:            db,
		pluginManager: pluginManager,
		themeManager:  themeManager,
		startTime:     time.Now(),
	}
}

func (d *DashboardManager) GetDashboardData() (*DashboardData, error) {
	// Get system statistics
	stats, err := d.getSystemStats()
	if err != nil {
		return nil, err
	}

	// Get recent activity
	activity, err := d.getRecentActivity()
	if err != nil {
		return nil, err
	}

	// Get system information
	systemInfo := d.getSystemInfo()

	// Get plugin status
	pluginStatus := d.getPluginStatus()

	return &DashboardData{
		Stats:          *stats,
		RecentActivity: activity,
		SystemInfo:     systemInfo,
		PluginStatus:   pluginStatus,
	}, nil
}

func (d *DashboardManager) getSystemStats() (*SystemStats, error) {
	ctx := context.Background()

	// Count total users
	userCollection := d.db.Collection("users")
	totalUsers, err := userCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Count active users (logged in within last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	activeUsers, err := userCollection.CountDocuments(ctx, bson.M{
		"last_login_at": bson.M{"$gte": thirtyDaysAgo},
	})
	if err != nil {
		return nil, err
	}

	// Get plugin statistics
	allPlugins := d.pluginManager.GetAllPlugins()
	totalPlugins := len(allPlugins)
	activePlugins := 0

	// Count active plugins from database
	pluginCollection := d.db.Collection("plugins")
	activePluginCount, err := pluginCollection.CountDocuments(ctx, bson.M{"is_active": true})
	if err == nil {
		activePlugins = int(activePluginCount)
	}

	// Get theme statistics
	allThemes := d.themeManager.GetAllThemes()
	totalThemes := len(allThemes)
	activeTheme := ""
	if activeThemeName := d.themeManager.GetActiveTheme(); activeThemeName != "" {
		activeTheme = activeThemeName
	}

	// Calculate uptime
	uptime := time.Since(d.startTime)
	uptimeStr := formatDuration(uptime)

	return &SystemStats{
		TotalUsers:    totalUsers,
		ActiveUsers:   activeUsers,
		TotalPlugins:  totalPlugins,
		ActivePlugins: activePlugins,
		TotalThemes:   totalThemes,
		ActiveTheme:   activeTheme,
		SystemUptime:  uptimeStr,
	}, nil
}

func (d *DashboardManager) getRecentActivity() ([]Activity, error) {
	// In a real implementation, you would query an activity log collection
	// For now, we'll return mock data
	return []Activity{
		{
			ID:          "act_001",
			Type:        "plugin",
			Description: "Plugin 'content-manager' was activated",
			User:        "admin",
			Timestamp:   time.Now().Add(-2 * time.Hour),
			Status:      "success",
		},
		{
			ID:          "act_002",
			Type:        "user",
			Description: "New user registration: john_doe",
			User:        "system",
			Timestamp:   time.Now().Add(-4 * time.Hour),
			Status:      "success",
		},
		{
			ID:          "act_003",
			Type:        "theme",
			Description: "Theme 'modern' was activated",
			User:        "admin",
			Timestamp:   time.Now().Add(-6 * time.Hour),
			Status:      "success",
		},
	}, nil
}

func (d *DashboardManager) getSystemInfo() SystemInfo {
	// Test database connection
	dbStatus := "connected"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := d.db.Client.Ping(ctx, nil); err != nil {
		dbStatus = "disconnected"
	}

	return SystemInfo{
		Version:        "1.0.0",
		GoVersion:      "1.21",
		StartTime:      d.startTime,
		DatabaseStatus: dbStatus,
		Environment:    "development", // This should come from config
	}
}

func (d *DashboardManager) getPluginStatus() []PluginStatus {
	var status []PluginStatus

	allPlugins := d.pluginManager.GetAllPlugins()
	for _, plugin := range allPlugins {
		info := plugin.GetInfo()

		pluginStatus := PluginStatus{
			Name:     info.Name,
			Version:  info.Version,
			Status:   "active",
			LoadTime: "< 1ms", // In a real implementation, you'd track this
		}

		status = append(status, pluginStatus)
	}

	return status
}

func formatDuration(d time.Duration) string {
	if d < time.Hour {
		return d.Round(time.Minute).String()
	}
	if d < 24*time.Hour {
		return d.Round(time.Hour).String()
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
