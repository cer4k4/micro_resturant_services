package api

import (
	"log"
	"order/models"

	httpReporter "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/labstack/echo/v4"
	"github.com/openzipkin/zipkin-go"
)

func Middleware(e *echo.Echo, configServer models.ServiceConfig) {
	reporter := httpReporter.NewReporter(configServer.ZipkinEndpoint)
	endpoint, _ := zipkin.NewEndpoint("CreateDelivery", "localhost:"+configServer.ServerPort)
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Println("tracer in Middleware", err)
	}
	// Middleware to add Zipkin tracing
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Extract trace context from request
			c.Set("tracer", tracer)
			return next(c)
		}
	})
}
