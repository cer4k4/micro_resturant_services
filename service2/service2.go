package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/openzipkin/zipkin-go"
	"github.com/openzipkin/zipkin-go/model"
	"github.com/openzipkin/zipkin-go/propagation/b3"
	kafkaReporter "github.com/openzipkin/zipkin-go/reporter/http"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"ame"`
}

func main() {
	// Initialize Zipkin tracer
	// var stlist []string
	// stlist = append(stlist, "localhost:9092")
	reporter := kafkaReporter.NewReporter("http://localhost:9411/api/v2/spans")
	// if err != nil {
	// 	log.Fatalf("Failed create kafka reporter: %v", err)
	// }

	endpoint, _ := zipkin.NewEndpoint("service2", "localhost:8081")
	tracer, _ := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/process", func(c echo.Context) error {
		// Extract tracing context from HTTP headers
		extractor := b3.ExtractHTTP(c.Request())
		spanContext, err := extractor()
		if err != nil && err != b3.ErrEmptyContext {
			log.Printf("Failed to extract span context: %v", err)
		}

		log.Println("service2", spanContext.TraceID, spanContext.ParentID, spanContext.ID)
		// Start a new span with the extracted context
		span := tracer.StartSpan("process-user", zipkin.Kind(model.Client), zipkin.Parent(*spanContext))
		defer span.Finish()

		// Parse the user data from the request body
		var user User
		if err := c.Bind(&user); err != nil {
			span.Tag("error", err.Error())
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid user data"})
		}
		log.Printf("Saving user: %+v", user)
		spanProducer := tracer.StartSpan("start-producer", zipkin.Kind(model.Producer), zipkin.Parent(span.Context()))
		defer spanProducer.Finish()
		spanProducer.Tag("produce", "hhhhhhh")

		spanConsumer := tracer.StartSpan("finish-producer", zipkin.Kind(model.Consumer), zipkin.Parent(spanProducer.Context()))
		defer spanConsumer.Finish()
		spanConsumer.Tag("consumer", "dfdssadfg")

		spanConsumer2 := tracer.StartSpan("finish-producer2", zipkin.Kind(model.Consumer), zipkin.Parent(spanConsumer.Context()))
		defer spanConsumer2.Finish()
		spanConsumer2.Tag("controller", "sdkjfsdkjhfkj")

		spanConsumer3 := tracer.StartSpan("finish-producer3", zipkin.Kind(model.Consumer), zipkin.Parent(spanConsumer.Context()))
		defer spanConsumer3.Finish()
		spanConsumer3.Tag("database", "djfdjkfgjjjjj")

		return c.JSON(http.StatusOK, map[string]string{"message": "User processed successfully"})
	})

	e.Logger.Fatal(e.Start(":8081"))
}
