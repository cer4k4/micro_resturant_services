package api

import (
	"order/databases"
	"order/models"
	"order/web"

	"github.com/labstack/echo/v4"
)

func Routes(data databases.MongoDBRepository, e *echo.Echo, configServer models.ServiceConfig) {
	handler := web.NewHandler(data, configServer)
	// Routes
	e.POST("/orders", handler.CreateOrder)
	e.GET("/orders", handler.ListOrders)
}
