package main

import (
	"delivery/api"
	"delivery/configs"
	"delivery/databases"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	configServer := configs.Loader()
	// Initialize Zipkin tracer
	repository := databases.NewMongoDB(configServer.MongoClient)
	e := echo.New()
	api.Middleware(e, configServer)
	api.Routes(repository, e)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Start server
	log.Fatal(e.Start(":" + configServer.ServerPort))
}
