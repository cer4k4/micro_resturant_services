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
	serverPort := os.Getenv("SERVER_PORT")

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

	endpoint, _ := zipkin.NewEndpoint("restaurant-service", "localhost:"+serverPort)
	tracer, _ = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/restaurants", createRestaurant)
	e.GET("/restaurants", listRestaurants)

	// Start server
	log.Fatal(e.Start(":" + serverPort))
}

func createRestaurant(c echo.Context) error {
	span := tracer.StartSpan("createRestaurant")
	defer span.Finish()

	var restaurant Restaurant
	if err := c.Bind(&restaurant); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	collection := mongoClient.Database("restaurantdb").Collection("restaurants")
	result, err := collection.InsertOne(context.TODO(), restaurant)
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create restaurant"})
	}

	restaurant.ID = result.InsertedID.(primitive.ObjectID)
	return c.JSON(http.StatusCreated, restaurant)
}

func listRestaurants(c echo.Context) error {
	span := tracer.StartSpan("listRestaurants")
	defer span.Finish()

	collection := mongoClient.Database("restaurantdb").Collection("restaurants")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch restaurants"})
	}
	defer cursor.Close(context.TODO())

	var restaurants []Restaurant
	for cursor.Next(context.TODO()) {
		var restaurant Restaurant
		if err := cursor.Decode(&restaurant); err != nil {
			span.Tag("error", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse restaurant data"})
		}
		restaurants = append(restaurants, restaurant)
	}

	return c.JSON(http.StatusOK, restaurants)
}
