package web

import (
	"log"
	"net/http"
	"resturant/databases"
	"resturant/models"

	"github.com/labstack/echo/v4"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
)

type Handler interface {
	CreateRestaurant(c echo.Context) error
	ListRestaurant(c echo.Context) error
}

type handler struct {
	db databases.MongoDBRepository
}

func NewHandler(db databases.MongoDBRepository) handler {
	return handler{db}
}

func (h *handler) CreateRestaurant(c echo.Context) error {
	tracer := c.Get("tracer").(*zipkin.Tracer)
	span := tracer.StartSpan("Create Restaurant Handler")
	defer span.Finish()

	var restaurant models.Restaurant
	if err := c.Bind(&restaurant); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	result, err := h.db.Create(c, restaurant)
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Can Not Create Resturant"})
	}
	return c.JSON(http.StatusCreated, result)
}

func (h *handler) ListRestaurants(c echo.Context) error {
	tracer := c.Get("tracer").(*zipkin.Tracer)
	extractor := b3.ExtractHTTP(c.Request())
	spanContext, err := extractor()
	if err != nil {
		log.Println("extractor faild", err)
	}
	span := tracer.StartSpan("List Restaurants Handler", zipkin.Kind(model.Client), zipkin.Parent(*spanContext))
	defer span.Finish()

	resturants, err := h.db.GetAll(c)
	if err != nil {
		span.Tag("error", err.Error())
		c.JSON(http.StatusBadRequest, map[string]string{"error": "Can Not Get All Resturant"})
	}
	return c.JSON(http.StatusOK, resturants)
}
