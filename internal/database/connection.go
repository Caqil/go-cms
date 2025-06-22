package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func Connect(mongoURL, dbName string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &DB{
		Client:   client,
		Database: client.Database(dbName),
	}, nil
}

func (db *DB) Disconnect(ctx context.Context) error {
	return db.Client.Disconnect(ctx)
}

func (db *DB) Collection(name string) *mongo.Collection {
	return db.Database.Collection(name)
}
