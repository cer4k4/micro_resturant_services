package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	httpReporter "github.com/openzipkin/zipkin-go/reporter/http"
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

var (
	mongoClient *mongo.Client
	tracer      *zipkin.Tracer
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mongoURI := os.Getenv("MONGODB_URI")
	zipkinEndpoint := os.Getenv("ZIPKIN_ENDPOINT")
	serverPort := os.Getenv("ORDER_SERVICE_PORT")

	// Initialize MongoDB client
	var err error
	mongoClient, err = mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.TODO())

	reporter := httpReporter.NewReporter(zipkinEndpoint)

	endpoint, _ := zipkin.NewEndpoint("order-service", "localhost:"+serverPort)
	tracer, _ = zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/orders", createOrder)
	e.GET("/orders", listOrders)

	// Start server
	log.Fatal(e.Start(":" + serverPort))
}

func createOrder(c echo.Context) error {
	span := tracer.StartSpan("createOrder", zipkin.Kind(model.Server))
	defer span.Finish()

	req, _ := http.NewRequest("GET", "http://localhost:8080/restaurants", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	b3.InjectHTTP(req)(span.Context())
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == 400 {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to call Service 2"})
	}
	defer resp.Body.Close()

	var restaurantResponse []Restaurant

	if err := json.NewDecoder(resp.Body).Decode(&restaurantResponse); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse restaurant response"})
	}

	var order Order

	if err := c.Bind(&order); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	order.RestaurantID = restaurantResponse[0].ID.String()
	order.Items = append(order.Items, OrderItem{MenuID: restaurantResponse[0].Menu[0].ID.String(), Quantity: 2, UnitPrice: restaurantResponse[0].Menu[0].Price})
	order.Status = "Pending"

	collection := mongoClient.Database("orderdb").Collection("orders")
	result, err := collection.InsertOne(context.TODO(), order)
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create order"})
	}

	order.ID = result.InsertedID.(primitive.ObjectID)
	return c.JSON(http.StatusCreated, order)
}

func listOrders(c echo.Context) error {
	extractor := b3.ExtractHTTP(c.Request())
	parentspanContext, err := extractor()
	if err != nil {
		log.Println("faild extractor from delivery", err)
	}
	span := tracer.StartSpan("listOrders", zipkin.Kind(model.Client), zipkin.Parent(*parentspanContext))
	defer span.Finish()

	collection := mongoClient.Database("orderdb").Collection("orders")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch orders"})
	}
	defer cursor.Close(context.TODO())

	var orders []Order
	for cursor.Next(context.TODO()) {
		var order Order
		if err := cursor.Decode(&order); err != nil {
			span.Tag("error", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse order data"})
		}
		orders = append(orders, order)
	}

	return c.JSON(http.StatusOK, orders)
}
