// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-cms/internal/config"
	"go-cms/internal/database"
	"go-cms/internal/database/migration"
	"go-cms/internal/plugins"
	"go-cms/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	// Initialize database connection
	db, err := database.Connect(cfg.MongoURI, cfg.DatabaseName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Disconnect(context.Background())

	// Run database migrations
	log.Println("Checking for database migrations...")
	migrationManager := migration.NewManager(db)
	if err := migrationManager.Run(); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// Initialize plugin manager
	pluginManager := plugins.NewManager()
	if err := pluginManager.LoadPlugins(cfg.PluginsDir); err != nil {
		log.Printf("Warning: Failed to load some plugins: %v", err)
	}

	// // Initialize theme manager
	// themeManager := themes.NewManager(cfg.ThemePath, db)
	// if err := themeManager.LoadThemes(); err != nil {
	// 	log.Printf("Warning: Failed to load themes: %v", err)
	// }

	// Setup Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router with dependencies
	r := router.Setup(&router.Dependencies{
		Config:        cfg,
		Database:      db,
		PluginManager: pluginManager,
		//ThemeManager:  themeManager,
	})

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("🚀 Server started on port %s", cfg.Port)
	log.Printf("📊 Admin interface available at: http://localhost:%s/admin", cfg.Port)
	log.Printf("🔑 Default admin credentials: admin@example.com / admin123")
	log.Printf("⚠️  Remember to change the default admin password!")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
