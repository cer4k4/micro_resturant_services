package api

import (
	"resturant/databases"
	"resturant/models"
	"resturant/web"

	"github.com/labstack/echo/v4"
)

func Routes(data databases.MongoDBRepository, e *echo.Echo, configServer models.ServiceConfig) {
	handler := web.NewHandler(data, configServer)
	// Routes
	e.POST("/restaurants", handler.CreateRestaurant)
	e.GET("/restaurants", handler.ListRestaurants)
}
