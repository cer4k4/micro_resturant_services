package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	kafkaReporter "github.com/openzipkin/zipkin-go/reporter/http"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User2 struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	// Initialize Zipkin tracer
	// var stlist []string
	// stlist = append(stlist, "localhost:9092")
	reporter := kafkaReporter.NewReporter("http://localhost:9411/api/v2/spans")
	// if err != nil {
	// 	log.Fatalf("Failed to create Kafka reporter: %v", err)
	// }
	defer reporter.Close()

	endpoint, _ := zipkin.NewEndpoint("service1", "localhost:8080")
	tracer, _ := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/call-service2", func(c echo.Context) error {
		// Start a new trace
		span := tracer.StartSpan("call-service2")
		defer span.Finish()

		// Create a request to Service 2
		user := User{ID: "123"}
		userJSON, _ := json.Marshal(user)

		req, _ := http.NewRequest("POST", "http://localhost:8081/process", bytes.NewBuffer(userJSON))
		req.Header.Set("Content-Type", "application/json")

		// Inject tracing context using B3 propagation

		b3.InjectHTTP(req)(span.Context())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode == 400 {
			span.Tag("error", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to call Service 2"})
		}
		defer resp.Body.Close()

		return c.JSON(http.StatusOK, map[string]string{"message": "Request to Service 2 successful"})
	})

	e.Logger.Fatal(e.Start(":8080"))
}
