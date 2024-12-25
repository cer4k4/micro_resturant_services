package databases

import (
	"context"
	"resturant/models"

	"github.com/labstack/echo/v4"
	"github.com/openzipkin/zipkin-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBRepository interface {
	Create(ctx echo.Context, resturant models.Restaurant) (models.Restaurant, error)
	GetAll(ctx echo.Context) ([]models.Restaurant, error)
}

type mongoDB struct {
	db *mongo.Client
}

func NewMongoDB(db *mongo.Client) MongoDBRepository {
	return &mongoDB{db}

}

func (m *mongoDB) Create(ctx echo.Context, resturant models.Restaurant) (models.Restaurant, error) {
	tracer := ctx.Get("tracer").(*zipkin.Tracer)
	span := ctx.Get("span").(zipkin.Span)
	spanDB := tracer.StartSpan("Database_Create_Resturant", zipkin.Parent(span.Context()))
	defer spanDB.Finish()
	collection := m.db.Database("restaurantdb").Collection("restaurants")

	result, err := collection.InsertOne(context.TODO(), resturant)
	if err != nil {
		span.Tag("error", err.Error())
		return models.Restaurant{}, err
	}

	resturant.ID = result.InsertedID.(primitive.ObjectID)
	return resturant, nil
}

func (m *mongoDB) GetAll(ctx echo.Context) ([]models.Restaurant, error) {
	tracer := ctx.Get("tracer").(*zipkin.Tracer)
	span := ctx.Get("span").(zipkin.Span)
	spanDB := tracer.StartSpan("Database_List_Resturant", zipkin.Parent(span.Context()))
	defer spanDB.Finish()

	collection := m.db.Database("restaurantdb").Collection("restaurants")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		span.Tag("error", err.Error())
		return []models.Restaurant{}, err
	}
	defer cursor.Close(context.TODO())

	var restaurants []models.Restaurant
	for cursor.Next(context.TODO()) {
		var restaurant models.Restaurant
		if err := cursor.Decode(&restaurant); err != nil {
			span.Tag("error", err.Error())
			return []models.Restaurant{}, err
		}
		restaurants = append(restaurants, restaurant)
	}
	return restaurants, nil
}
