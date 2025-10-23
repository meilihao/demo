package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
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

	go everLoop(ctx)
	router.Run(":5000")
}

func everLoop(ctx context.Context) {
LOOP:
	for {
		select {
		case <-ctx.Done():
			// 收到信号退出无限循环
			break LOOP
		default:
			// 用一个 sleep 模拟业务逻辑
			time.Sleep(time.Second * 10)
		}
	}
}
