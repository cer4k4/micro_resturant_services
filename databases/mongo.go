package databases

import (
	"context"
	"order/models"

	"github.com/labstack/echo/v4"
	"github.com/openzipkin/zipkin-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDBRepository interface {
	Create(ctx echo.Context, order models.Order) (models.Order, error)
	GetAll(ctx echo.Context) ([]models.Order, error)
}

type mongoDB struct {
	db *mongo.Client
}

func NewMongoDB(db *mongo.Client) MongoDBRepository {
	return &mongoDB{db}

}

func (m *mongoDB) Create(ctx echo.Context, order models.Order) (models.Order, error) {
	tracer := ctx.Get("tracer").(*zipkin.Tracer)
	span := ctx.Get("span").(zipkin.Span)
	spanDB := tracer.StartSpan("Database_Create_Order", zipkin.Parent(span.Context()))
	defer spanDB.Finish()
	collection := m.db.Database("orderdb").Collection("orders")
	result, err := collection.InsertOne(context.TODO(), order)
	if err != nil {
		spanDB.Tag("error", err.Error())
		return models.Order{}, err
	}
	spanDB.Tag("MongoQuery", "db.orders.insertOne({_id: ObjectId(), items: [{menu_id: '1',quantity: 1,unit_price: 10.99}],status: 'pending'})")
	order.ID = result.InsertedID.(primitive.ObjectID)
	return order, nil
}

func (m *mongoDB) GetAll(ctx echo.Context) ([]models.Order, error) {
	tracer := ctx.Get("tracer").(*zipkin.Tracer)
	span := ctx.Get("span").(zipkin.Span)
	spanDB := tracer.StartSpan("Database_List_Order", zipkin.Parent(span.Context()))
	defer spanDB.Finish()
	collection := m.db.Database("orderdb").Collection("orders")
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		spanDB.Tag("error", err.Error())
		return []models.Order{}, err
	}
	defer cursor.Close(context.TODO())

	var orders []models.Order
	for cursor.Next(context.TODO()) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			spanDB.Tag("error", err.Error())
			return []models.Order{}, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
