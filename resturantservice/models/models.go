package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Restaurant struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Address  string             `bson:"address" json:"address"`
	Cuisine  string             `bson:"cuisine" json:"cuisine"`
	Menu     []Menu             `bson:"menu" json:"menu"`
	IsActive bool               `bson:"is_active" json:"is_active"`
}

type Menu struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Price       float64            `bson:"price" json:"price"`
	Category    string             `bson:"category" json:"category"`
}

type ServiceConfig struct {
	ZipkinEndpoint string
	ServerPort     string
	MongoClient    *mongo.Client
}
