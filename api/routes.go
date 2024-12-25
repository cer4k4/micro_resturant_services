package api

import (
	"resturant/databases"
	"resturant/web"

	"github.com/labstack/echo/v4"
)

func Routes(data databases.MongoDBRepository, e *echo.Echo) {
	handler := web.NewHandler(data)
	// Routes
	e.POST("/restaurants", handler.CreateRestaurant)
	e.GET("/restaurants", handler.ListRestaurants)
}
