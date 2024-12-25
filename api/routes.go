package api

import (
	"order/databases"
	"order/web"

	"github.com/labstack/echo/v4"
)

func Routes(data databases.MongoDBRepository, e *echo.Echo) {
	handler := web.NewHandler(data)
	// Routes
	e.POST("/orders", handler.CreateOrder)
	e.GET("/orders", handler.ListOrders)
}
