package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Order struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RestaurantID string             `bson:"restaurant_id" json:"restaurant_id"`
	Items        []OrderItem        `bson:"items,omitempty" json:"items"`
	Status       string             `bson:"status" json:"status"`
}

type OrderItem struct {
	MenuID    string  `bson:"menu_id" json:"menu_id"`
	Quantity  int     `bson:"quantity" json:"quantity"`
	UnitPrice float64 `bson:"unit_price" json:"unit_price"`
}

type Delivery struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID   string             `bson:"order_id" json:"order_id"`
	DriverID  string             `bson:"driver_id" json:"driver_id"`
	Status    string             `bson:"status" json:"status"`
	Timestamp string             `bson:"timestamp" json:"timestamp"`
}

type ServiceConfig struct {
	ZipkinEndpoint string
	ServerPort     string
	MongoClient    *mongo.Client
}
