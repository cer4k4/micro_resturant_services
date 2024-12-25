package configs

import (
	"context"
	"log"
	"order/models"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Loader() (config models.ServiceConfig) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongoURI := os.Getenv("MONGODB_URI")
	config.ZipkinEndpoint = os.Getenv("ZIPKIN_ENDPOINT")
	config.ServerPort = os.Getenv("ORDER_SERVICE_PORT")

	// Initialize MongoDB client
	var err error
	config.MongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	return
}
