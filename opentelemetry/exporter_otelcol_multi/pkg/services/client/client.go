// see https://github.com/open-telemetry/opentelemetry-go-contrib/blob/master/instrumentation/google.golang.org/grpc/otelgrpc/example/client/main.go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"time"

	"distributed-tracing-otel/pkg/weatherpb"

	"github.com/meilihao/golib/v2"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	otelEndpoint string
)

func init() {
	flag.StringVar(&otelEndpoint, "otel_endpoint", "localhost:4317", "otel endpoint")
}

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdownFn, err := golib.SetupOTelSDK(ctx, otelEndpoint, "client")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdownFn(ctx)

	tracer := otel.Tracer("client")

	cc, err := grpc.NewClient("localhost:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()), // 在将请求发送到gRPC服务器之前，此拦截器会将span信息添加到上下文中
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	)

	if err != nil {
		log.Fatalf("Error connecting: %v", err)
	}
	defer cc.Close()

	c := weatherpb.NewWeatherServiceClient(cc)
	getCurrentWeather(c, tracer)
}

func getCurrentWeather(c weatherpb.WeatherServiceClient, tracer trace.Tracer) {
	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	// <namespace>_an_important_metric{labelA="chocolate",labelB="raspberry",labelC="vanilla"} 2
	commonLabels := []attribute.KeyValue{
		attribute.String("labelA", "chocolate"),
		attribute.String("labelB", "raspberry"),
		attribute.String("labelC", "vanilla"),
	}

	ctx, span := tracer.Start(context.Background(), "GetCurrentWeather", trace.WithAttributes(commonLabels...))
	defer span.End()

	meter := otel.Meter("test-meter")

	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/main/examples/otel-collector/main.go
	counter, err := meter.Int64Counter(
		"an_important_metric",
		metric.WithDescription("Measures the cumulative epicness of the app"),
	)
	if err != nil {
		log.Fatalf("Failed to create the instrument: %v", err)
	}
	counter.Add(ctx, 1, metric.WithAttributes(commonLabels...))

	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	logger := otelslog.NewLogger("client")

	req := &weatherpb.WeatherRequest{
		Location: "dublin",
	}

	res, err := c.GetCurrentWeather(ctx, req)
	if err != nil {
		logger.ErrorContext(ctx, "GetCurrentWeather", "err", err)
		span.RecordError(err)
		return
	}
	logger.InfoContext(ctx, "GetCurrentWeather", "result", res)

	span.AddEvent("Response", trace.WithAttributes(
		attribute.String("condition", res.Condition),
		attribute.Float64("temperature", res.Temperature),
	))
}
