package main

import (
	"log"

	"distributed-tracing-otel/pkg/lib"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/trace"

	"github.com/gin-gonic/gin"
)

func Temperature(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "Temperature")
	defer span.End()

	span.AddEvent("header", trace.WithAttributes(
		label.Any("data", c.Request.Header),
	))

	tmp := 1.0

	span.AddEvent("Response", trace.WithAttributes(
		label.Float64("return temperature", tmp),
	))

	c.JSON(200, gin.H{
		"temperatureC": tmp,
	})
}

var (
	tracer trace.Tracer
)

func main() {
	shutdownFn, err := lib.InitOTEL("openhello.net:55680", "temperature")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdownFn()

	tracer = otel.Tracer("temperature-tracer")

	router := gin.New()
	router.Use(otelgin.Middleware("gin-middleware"))
	router.GET("/WeatherForecast", Temperature)

	router.Run(":5000")
}
