package api

import (
	"delivery/databases"
	"delivery/web"

	"github.com/labstack/echo/v4"
)

func Routes(data databases.MongoDBRepository, e *echo.Echo) {
	handler := web.NewHandler(data)
	// Routes
	e.POST("/deliveries", handler.CreateDelivery)
	e.GET("/deliveries", handler.ListDeliveries)
}
