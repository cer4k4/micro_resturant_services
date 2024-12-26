package api

import (
	"delivery/databases"
	"delivery/models"
	"delivery/web"

	"github.com/labstack/echo/v4"
)

func Routes(data databases.MongoDBRepository, e *echo.Echo, configServer models.ServiceConfig) {
	handler := web.NewHandler(data, configServer)
	// Routes
	e.POST("/deliveries", handler.CreateDelivery)
	e.GET("/deliveries", handler.ListDeliveries)
}
