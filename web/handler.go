package web

import (
	"encoding/json"
	"log"
	"net/http"
	"order/databases"
	"order/models"

	"github.com/labstack/echo/v4"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	httpReporter "github.com/openzipkin/zipkin-go/reporter/http"
)

type Handler interface {
	CreateOrder(c echo.Context) error
	ListOrders(c echo.Context) error
}

type handler struct {
	db           databases.MongoDBRepository
	configServer models.ServiceConfig
}

func NewHandler(db databases.MongoDBRepository, configServer models.ServiceConfig) handler {
	return handler{db, configServer}
}

func (h *handler) CreateOrder(c echo.Context) error {
	reporter := httpReporter.NewReporter(h.configServer.ZipkinEndpoint)
	endpoint, _ := zipkin.NewEndpoint("Create Order Service", "localhost:"+h.configServer.ServerPort)
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	c.Set("tracer", tracer)
	if err != nil {
		log.Println("tracer in Middleware", err)
	}
	span := tracer.StartSpan("Create Order Handler", zipkin.Kind(model.Server))
	defer span.Finish()
	c.Set("span", span)
	req, err := http.NewRequest("GET", "http://localhost:8081/restaurants", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	b3.InjectHTTP(req)(span.Context())
	client := &http.Client{}
	if err != nil {
		log.Println("error after send request to service ", err)
		span.Tag("error", err.Error())
	}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode == 400 {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to call Service 2"})
	}
	defer resp.Body.Close()

	var restaurantResponse []models.Restaurant

	if err := json.NewDecoder(resp.Body).Decode(&restaurantResponse); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse restaurant response"})
	}

	var order models.Order

	if err := c.Bind(&order); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	order.RestaurantID = restaurantResponse[0].ID.String()
	order.Items = append(order.Items, models.OrderItem{MenuID: restaurantResponse[0].Menu[0].ID.String(), Quantity: 2, UnitPrice: restaurantResponse[0].Menu[0].Price})
	order.Status = "Pending"

	result, err := h.db.Create(c, order)
	if err != nil {
		log.Println("error failed create order ", err)
		span.Tag("error", err.Error())
	}
	span.Tag(string(zipkin.TagHTTPMethod), "POST")
	span.Tag(string(zipkin.TagHTTPStatusCode), http.StatusText(http.StatusCreated))
	span.Tag(string(zipkin.TagHTTPUrl), "/orders")
	cs := c.QueryParams()["pretty"]
	log.Println(cs)
	//span.Tag("http-response",)
	return c.JSON(http.StatusCreated, result)
}

func (h *handler) ListOrders(c echo.Context) error {
	reporter := httpReporter.NewReporter(h.configServer.ZipkinEndpoint)
	endpoint, _ := zipkin.NewEndpoint("List Order", "localhost:"+h.configServer.ServerPort)
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Println("tracer in Middleware", err)
	}

	extractor := b3.ExtractHTTP(c.Request())
	parentspanContext, err := extractor()
	if err != nil {
		log.Println("faild extractor from delivery", err)
	}
	span := tracer.StartSpan("List Orders Handler", zipkin.Kind(model.Client), zipkin.Parent(*parentspanContext))
	defer span.Finish()
	c.Set("span", span)
	c.Set("tracer", tracer)
	orders, err := h.db.GetAll(c)
	if err != nil {
		span.Tag("error", "Faild_Get_All_Order_Handler")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Faild Get orders"})
	}
	return c.JSON(http.StatusOK, orders)
}
