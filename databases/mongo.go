package databases

import (
	"context"
	"delivery/models"

	"github.com/labstack/echo/v4"
	"github.com/openzipkin/zipkin-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBRepository interface {
	Create(ctx echo.Context, delivery models.Delivery) (models.Delivery, error)
	GetAll(ctx echo.Context) ([]models.Delivery, error)
}

type mongoDB struct {
	db *mongo.Client
}

func NewMongoDB(db *mongo.Client) MongoDBRepository {
	return &mongoDB{db}

}

func (m *mongoDB) Create(ctx echo.Context, delivery models.Delivery) (models.Delivery, error) {
	tracer := ctx.Get("tracer").(*zipkin.Tracer)
	span := ctx.Get("span").(zipkin.Span)
	spanDB := tracer.StartSpan("Database_Create_Delivery", zipkin.Parent(span.Context()))
	defer spanDB.Finish()
	collection := m.db.Database("deliverydb").Collection("deliveries")
	result, err := collection.InsertOne(context.TODO(), delivery)
	if err != nil {
		spanDB.Tag("error", err.Error())
		return models.Delivery{}, err
	}
	delivery.ID = result.InsertedID.(primitive.ObjectID)
	return delivery, nil
}

func (m *mongoDB) GetAll(ctx echo.Context) ([]models.Delivery, error) {
	tracer := ctx.Get("tracer").(*zipkin.Tracer)
	span := ctx.Get("span").(zipkin.Span)
	spanDB := tracer.StartSpan("Database_List_Delivery", zipkin.Parent(span.Context()))
	defer spanDB.Finish()
	collection := m.db.Database("deliverydb").Collection("deliveries")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		spanDB.Tag("error", err.Error())
		return []models.Delivery{}, err
	}
	defer cursor.Close(context.TODO())

	var deliveries []models.Delivery
	for cursor.Next(context.TODO()) {
		var delivery models.Delivery
		if err := cursor.Decode(&delivery); err != nil {
			spanDB.Tag("error", err.Error())
			return []models.Delivery{}, err
		}
		deliveries = append(deliveries, delivery)
	}
	return deliveries, nil
}
