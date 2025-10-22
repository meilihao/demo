package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/gin-gonic/gin"
	"github.com/meilihao/golib/v2"
)

func Temperature(c *gin.Context) {
	_, span := tracer.Start(c.Request.Context(), "Temperature")
	defer span.End()

	span.AddEvent("header", trace.WithAttributes(
		attribute.String("data", golib.DumpToJsonString(c.Request.Header)),
	))

	tmp := 1.0

	span.AddEvent("Response", trace.WithAttributes(
		attribute.Float64("return temperature", tmp),
	))

	c.JSON(200, gin.H{
		"temperatureC": tmp,
	})
}

var (
	otelEndpoint string
	tracer       trace.Tracer
)

func init() {
	flag.StringVar(&otelEndpoint, "otel_endpoint", "localhost:4317", "otel endpoint")
}

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdownFn, err := golib.SetupOTelSDK(ctx, otelEndpoint, "temperature")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdownFn(ctx)

	tracer = otel.Tracer("temperature-tracer")

	router := gin.New()
	router.Use(otelgin.Middleware("gin-middleware"))
	router.GET("/WeatherForecast", Temperature)

	router.Run(":5000")
}
