package models

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type PluginMetadata struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
    Name        string             `bson:"name" json:"name"`
    Version     string             `bson:"version" json:"version"`
    Description string             `bson:"description" json:"description"`
    Author      string             `bson:"author" json:"author"`
    Website     string             `bson:"website,omitempty" json:"website,omitempty"`
    Filename    string             `bson:"filename" json:"filename"`
    IsActive    bool               `bson:"is_active" json:"is_active"`
    Settings    []PluginSetting    `bson:"settings" json:"settings"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type PluginSetting struct {
    Key         string      `bson:"key" json:"key"`
    Label       string      `bson:"label" json:"label"`
    Type        string      `bson:"type" json:"type"` // text, number, boolean, select
    Value       interface{} `bson:"value" json:"value"`
    Description string      `bson:"description,omitempty" json:"description,omitempty"`
    Options     []string    `bson:"options,omitempty" json:"options,omitempty"`
    Required    bool        `bson:"required" json:"required"`
}

type PluginUpload struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
    Author      string `json:"author"`
}