package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/reporter/kafka"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Delivery struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID   string             `bson:"order_id" json:"order_id"`
	DriverID  string             `bson:"driver_id" json:"driver_id"`
	Status    string             `bson:"status" json:"status"`
	Timestamp string             `bson:"timestamp" json:"timestamp"`
}

var (
	mongoClient *mongo.Client
	tracer      *zipkin.Tracer
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongoURI := os.Getenv("MONGO_URI")
	//zipkinEndpoint := os.Getenv("ZIPKIN_ENDPOINT")
	serverPort := os.Getenv("DELIVERY_SERVICE_PORT")

	// Initialize MongoDB client
	var err error
	mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.TODO())

	// Initialize Zipkin tracer
	var addressKafka []string
	addressKafka = append(addressKafka, "localhost:9092")
	reporter, err := kafka.NewReporter(addressKafka)

	endpoint, _ := zipkin.NewEndpoint("delivery-service", "localhost:"+serverPort)
	tracer, _ = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/deliveries", createDelivery)
	e.GET("/deliveries", listDeliveries)

	// Start server
	log.Fatal(e.Start(":" + serverPort))
}

func createDelivery(c echo.Context) error {
	span := tracer.StartSpan("createDelivery")
	defer span.Finish()

	var delivery Delivery
	if err := c.Bind(&delivery); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	delivery.Status = "In Progress"
	collection := mongoClient.Database("deliverydb").Collection("deliveries")
	result, err := collection.InsertOne(context.TODO(), delivery)
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create delivery"})
	}

	delivery.ID = result.InsertedID.(primitive.ObjectID)
	return c.JSON(http.StatusCreated, delivery)
}

func listDeliveries(c echo.Context) error {
	span := tracer.StartSpan("listDeliveries")
	defer span.Finish()

	collection := mongoClient.Database("deliverydb").Collection("deliveries")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch deliveries"})
	}
	defer cursor.Close(context.TODO())

	var deliveries []Delivery
	for cursor.Next(context.TODO()) {
		var delivery Delivery
		if err := cursor.Decode(&delivery); err != nil {
			span.Tag("error", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse delivery data"})
		}
		deliveries = append(deliveries, delivery)
	}

	return c.JSON(http.StatusOK, deliveries)
}
