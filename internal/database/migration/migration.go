// internal/database/migration/migration.go
package migration

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"go-cms/internal/database"
	"go-cms/internal/database/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Migration represents a database migration
type Migration struct {
	Version     string
	Description string
	Up          func(db *database.DB) error
	Down        func(db *database.DB) error
}

// MigrationRecord represents a migration record in the database
type MigrationRecord struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Version     string             `bson:"version"`
	Description string             `bson:"description"`
	AppliedAt   time.Time          `bson:"applied_at"`
	Success     bool               `bson:"success"`
}

// Manager handles database migrations
type Manager struct {
	db         *database.DB
	migrations []Migration
}

// NewManager creates a new migration manager
func NewManager(db *database.DB) *Manager {
	return &Manager{
		db:         db,
		migrations: getMigrations(),
	}
}

// Run executes all pending migrations
func (m *Manager) Run() error {
	log.Println("Starting database migrations...")

	// Ensure migrations collection exists
	if err := m.ensureMigrationsCollection(); err != nil {
		return fmt.Errorf("failed to ensure migrations collection: %w", err)
	}

	// Get applied migrations
	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Sort migrations by version
	sort.Slice(m.migrations, func(i, j int) bool {
		return m.migrations[i].Version < m.migrations[j].Version
	})

	// Run pending migrations
	for _, migration := range m.migrations {
		if _, exists := applied[migration.Version]; exists {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %s: %s", migration.Version, migration.Description)

		if err := m.applyMigration(migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		log.Printf("Successfully applied migration %s", migration.Version)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// applyMigration applies a single migration
func (m *Manager) applyMigration(migration Migration) error {
	ctx := context.Background()

	// Start transaction if possible (for replica sets)
	session, err := m.db.Client.StartSession()
	if err != nil {
		// If sessions are not supported, continue without transaction
		if err := migration.Up(m.db); err != nil {
			return err
		}
		return m.recordMigration(migration, true)
	}
	defer session.EndSession(ctx)

	// Use transaction
	err = mongo.WithSession(ctx, session, func(sc mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		if err := migration.Up(m.db); err != nil {
			session.AbortTransaction(sc)
			return err
		}

		if err := m.recordMigration(migration, true); err != nil {
			session.AbortTransaction(sc)
			return err
		}

		return session.CommitTransaction(sc)
	})

	return err
}

// recordMigration records the migration in the database
func (m *Manager) recordMigration(migration Migration, success bool) error {
	collection := m.db.Collection("_migrations")

	record := MigrationRecord{
		Version:     migration.Version,
		Description: migration.Description,
		AppliedAt:   time.Now(),
		Success:     success,
	}

	_, err := collection.InsertOne(context.Background(), record)
	return err
}

// getAppliedMigrations returns a map of applied migration versions
func (m *Manager) getAppliedMigrations() (map[string]bool, error) {
	collection := m.db.Collection("_migrations")

	cursor, err := collection.Find(context.Background(), bson.M{"success": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	applied := make(map[string]bool)
	for cursor.Next(context.Background()) {
		var record MigrationRecord
		if err := cursor.Decode(&record); err != nil {
			return nil, err
		}
		applied[record.Version] = true
	}

	return applied, cursor.Err()
}

// ensureMigrationsCollection ensures the migrations collection exists
func (m *Manager) ensureMigrationsCollection() error {
	collection := m.db.Collection("_migrations")

	// Create index on version field
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "version", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		// Ignore error if index already exists
		if !mongo.IsDuplicateKeyError(err) {
			return err
		}
	}

	return nil
}

// getMigrations returns all available migrations
func getMigrations() []Migration {
	return []Migration{
		{
			Version:     "001_initial_setup",
			Description: "Create initial collections and indexes",
			Up:          migration001Up,
			Down:        migration001Down,
		},
		{
			Version:     "002_users_indexes",
			Description: "Create user collection indexes",
			Up:          migration002Up,
			Down:        migration002Down,
		},
		{
			Version:     "003_plugins_indexes",
			Description: "Create plugin collection indexes",
			Up:          migration003Up,
			Down:        migration003Down,
		},
		{
			Version:     "004_themes_indexes",
			Description: "Create theme collection indexes",
			Up:          migration004Up,
			Down:        migration004Down,
		},
		{
			Version:     "005_initial_data",
			Description: "Insert initial data",
			Up:          migration005Up,
			Down:        migration005Down,
		},
	}
}

// Migration 001: Initial setup
func migration001Up(db *database.DB) error {
	log.Println("Setting up initial database structure...")

	// Collections will be created automatically when first document is inserted
	// This migration just ensures they exist and have basic validation

	collections := []string{"users", "plugins", "themes", "theme_configs", "theme_backups", "theme_customizations"}

	for _, collName := range collections {
		// Check if collection exists
		collection := db.Collection(collName)

		// Create a temporary document to ensure collection exists, then remove it
		tempDoc := bson.M{"_temp": true}
		result, err := collection.InsertOne(context.Background(), tempDoc)
		if err != nil {
			return fmt.Errorf("failed to create collection %s: %w", collName, err)
		}

		// Remove the temporary document
		_, err = collection.DeleteOne(context.Background(), bson.M{"_id": result.InsertedID})
		if err != nil {
			return fmt.Errorf("failed to clean up temp document in %s: %w", collName, err)
		}

		log.Printf("Collection %s initialized", collName)
	}

	return nil
}

func migration001Down(db *database.DB) error {
	// In a real scenario, you might want to drop collections
	// For safety, we'll leave them as is
	log.Println("Migration 001 rollback - collections left intact")
	return nil
}

// Migration 002: Users indexes
func migration002Up(db *database.DB) error {
	log.Println("Creating user collection indexes...")

	collection := db.Collection("users")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "role", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create user indexes: %w", err)
	}

	log.Println("User indexes created successfully")
	return nil
}

func migration002Down(db *database.DB) error {
	collection := db.Collection("users")
	_, err := collection.Indexes().DropAll(context.Background())
	return err
}

// Migration 003: Plugins indexes
func migration003Up(db *database.DB) error {
	log.Println("Creating plugin collection indexes...")

	collection := db.Collection("plugins")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "version", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	}

	_, err := collection.Indexes().CreateMany(context.Background(), indexes)
	if err != nil {
		return fmt.Errorf("failed to create plugin indexes: %w", err)
	}

	log.Println("Plugin indexes created successfully")
	return nil
}

func migration003Down(db *database.DB) error {
	collection := db.Collection("plugins")
	_, err := collection.Indexes().DropAll(context.Background())
	return err
}

// Migration 004: Themes indexes
func migration004Up(db *database.DB) error {
	log.Println("Creating theme collection indexes...")

	// Themes collection
	themesCollection := db.Collection("themes")
	themeIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "version", Value: 1}},
		},
	}

	_, err := themesCollection.Indexes().CreateMany(context.Background(), themeIndexes)
	if err != nil {
		return fmt.Errorf("failed to create theme indexes: %w", err)
	}

	// Theme configs collection
	configsCollection := db.Collection("theme_configs")
	configIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "theme_name", Value: 1}, {Key: "config_key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "theme_name", Value: 1}},
		},
	}

	_, err = configsCollection.Indexes().CreateMany(context.Background(), configIndexes)
	if err != nil {
		return fmt.Errorf("failed to create theme config indexes: %w", err)
	}

	// Theme customizations collection
	customizationsCollection := db.Collection("theme_customizations")
	customizationIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "theme_name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: 1}},
		},
	}

	_, err = customizationsCollection.Indexes().CreateMany(context.Background(), customizationIndexes)
	if err != nil {
		return fmt.Errorf("failed to create theme customization indexes: %w", err)
	}

	log.Println("Theme indexes created successfully")
	return nil
}

func migration004Down(db *database.DB) error {
	collections := []string{"themes", "theme_configs", "theme_customizations"}
	for _, collName := range collections {
		collection := db.Collection(collName)
		_, err := collection.Indexes().DropAll(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

// Migration 005: Initial data
func migration005Up(db *database.DB) error {
	log.Println("Inserting initial data...")

	// Create default admin user if no users exist
	usersCollection := db.Collection("users")
	userCount, err := usersCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count users: %w", err)
	}

	if userCount == 0 {
		log.Println("Creating default admin user...")

		adminUser := models.User{
			Username:  "admin",
			Email:     "admin@example.com",
			Password:  "admin123", // This will be hashed
			Role:      "super_admin",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Hash the password
		if err := adminUser.HashPassword(); err != nil {
			return fmt.Errorf("failed to hash admin password: %w", err)
		}

		_, err := usersCollection.InsertOne(context.Background(), adminUser)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		log.Println("Default admin user created (email: admin@example.com, password: admin123)")
		log.Println("⚠️  IMPORTANT: Change the default admin password after first login!")
	}

	// Create default theme if no themes exist
	themesCollection := db.Collection("themes")
	themeCount, err := themesCollection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return fmt.Errorf("failed to count themes: %w", err)
	}

	if themeCount == 0 {
		log.Println("Creating default theme entry...")

		defaultTheme := models.ThemeMetadata{
			Name:        "default",
			Version:     "1.0.0",
			Description: "Default CMS theme",
			Author:      "CMS Team",
			IsActive:    true,
			UpdatedAt:   time.Now(),
		}

		_, err := themesCollection.InsertOne(context.Background(), defaultTheme)
		if err != nil {
			return fmt.Errorf("failed to create default theme: %w", err)
		}

		log.Println("Default theme entry created")
	}

	log.Println("Initial data setup completed")
	return nil
}

func migration005Down(db *database.DB) error {
	// Remove default data
	usersCollection := db.Collection("users")
	_, err := usersCollection.DeleteOne(context.Background(), bson.M{"username": "admin"})
	if err != nil {
		return err
	}

	themesCollection := db.Collection("themes")
	_, err = themesCollection.DeleteOne(context.Background(), bson.M{"name": "default"})
	return err
}
