package router

import (
	"go-cms/internal/admin"
	"go-cms/internal/auth"
	"go-cms/internal/config"
	"go-cms/internal/database"
	"go-cms/internal/middleware"
	"go-cms/internal/plugins"
	"go-cms/internal/themes"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Config        *config.Config
	Database      *database.DB
	PluginManager *plugins.Manager
	ThemeManager  *themes.Manager
}

func Setup(deps *Dependencies) *gin.Engine {
	r := gin.Default()

	// Set upload limit for plugin files (100MB)
	r.MaxMultipartMemory = 100 << 20

	// Set up plugin dependencies
	pluginDeps := &plugins.PluginDependencies{
		Database: deps.Database,
		Config:   deps.Config,
	}
	deps.PluginManager.SetDependencies(pluginDeps)

	// Middleware
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger())

	// Public routes
	public := r.Group("/api/v1")
	{
		// Auth routes
		authHandler := auth.NewHandler(deps.Database, deps.Config.JWTSecret)
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
		public.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(auth.JWTMiddleware(deps.Config.JWTSecret))
	{
		// User routes
		authHandler := auth.NewHandler(deps.Database, deps.Config.JWTSecret)
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)

		// Theme routes
		themeHandler := themes.NewHandler(deps.ThemeManager)
		protected.GET("/themes", themeHandler.GetAll)
		protected.GET("/themes/:name", themeHandler.GetTheme)
		protected.POST("/themes/:name/activate", auth.AdminRequired(), themeHandler.ActivateTheme)
	}

	// Admin routes
	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(auth.JWTMiddleware(deps.Config.JWTSecret))
	adminGroup.Use(auth.AdminRequired())
	{
		adminHandler := admin.NewHandler(deps.Database, deps.PluginManager, deps.ThemeManager)

		// Dashboard
		adminGroup.GET("/dashboard", adminHandler.GetDashboard)
		adminGroup.GET("/menu", adminHandler.GetMenu)

		// Plugin management
		adminGroup.GET("/plugins", adminHandler.GetPlugins)
		adminGroup.POST("/plugins/upload", adminHandler.UploadPlugin)
		adminGroup.POST("/plugins/:name/toggle", adminHandler.TogglePlugin)
		adminGroup.POST("/plugins/:name/reload", adminHandler.ReloadPlugin)
		adminGroup.DELETE("/plugins/:name", adminHandler.DeletePlugin)

		// Plugin settings
		adminGroup.GET("/plugins/:name/settings", adminHandler.GetPluginSettings)
		adminGroup.PUT("/plugins/:name/settings", adminHandler.UpdatePluginSettings)

		// System management
		adminGroup.GET("/system/info", adminHandler.GetSystemInfo)
		adminGroup.POST("/system/cleanup-cache", adminHandler.CleanupCache)
		adminGroup.POST("/system/hot-reload", adminHandler.HotReloadAll)
	}

	// Plugin routes - Store router reference for dynamic registration
	deps.PluginManager.SetRouter(protected)
	deps.PluginManager.RegisterRoutes(protected)

	// Static file serving for admin dashboard
	r.Static("/admin", deps.Config.AdminPath)
	r.Static("/themes", deps.Config.ThemePath)
	r.Static("/uploads", "./uploads") // For plugin assets

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "go-cms",
		})
	})

	return r
}
