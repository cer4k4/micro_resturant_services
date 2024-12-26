package web

import (
	"delivery/databases"
	"delivery/models"
	"encoding/json"
	"log"
	"net/http"

	httpReporter "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/labstack/echo/v4"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
)

type Handler interface {
	CreateDelivery(c echo.Context) error
	ListDeliveries(c echo.Context) error
}

type handler struct {
	db           databases.MongoDBRepository
	configServer models.ServiceConfig
}

func NewHandler(db databases.MongoDBRepository, configServer models.ServiceConfig) handler {
	return handler{db, configServer}
}

func (h *handler) CreateDelivery(c echo.Context) error {

	reporter := httpReporter.NewReporter(h.configServer.ZipkinEndpoint)
	endpoint, _ := zipkin.NewEndpoint("CreateDelivery", "localhost:"+h.configServer.ServerPort)
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Println("tracer in Middleware", err)
	}
	span := tracer.StartSpan("Create_Delivery_Handler", zipkin.Kind(model.Server))
	defer span.Finish()
	c.Set("span", span)
	c.Set("tracer", tracer)
	req, err := http.NewRequest("GET", "http://localhost:8082/orders", http.NoBody)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Println("error after send request to service ", err)
		span.Tag("error", err.Error())
	}
	b3.InjectHTTP(req)(span.Context())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to call restaurant service"})
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return c.JSON(resp.StatusCode, map[string]string{"error": "Restaurant service returned error"})
	}

	var orders []models.Order

	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to parse restaurant response"})
	}

	var delivery models.Delivery
	if err := c.Bind(&delivery); err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	delivery.OrderID = orders[0].ID.String()
	delivery.DriverID = "474747"
	delivery.Status = "In Progress"
	deliver, err := h.db.Create(c, delivery)

	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create delivery"})
	}
	return c.JSON(http.StatusCreated, deliver)
}

func (h *handler) ListDeliveries(c echo.Context) error {

	reporter := httpReporter.NewReporter(h.configServer.ZipkinEndpoint)
	endpoint, _ := zipkin.NewEndpoint("List Deliveries", "localhost:"+h.configServer.ServerPort)
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Println("tracer in Middleware", err)
	}
	span := tracer.StartSpan("List_Deliveries_Handler")
	defer span.Finish()
	c.Set("span", span)
	c.Set("tracer", tracer)
	list, err := h.db.GetAll(c)

	if err != nil {
		span.Tag("error", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch deliveries"})
	}

	return c.JSON(http.StatusOK, list)
}
