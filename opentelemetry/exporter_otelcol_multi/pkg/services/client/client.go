// see https://github.com/open-telemetry/opentelemetry-go-contrib/blob/master/instrumentation/google.golang.org/grpc/otelgrpc/example/client/main.go
package main

import (
	"context"
	"log"
	"time"

	"distributed-tracing-otel/pkg/lib"
	"distributed-tracing-otel/pkg/weatherpb"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	shutdownFn, err := lib.InitOTEL("openhello.net:4317", "client")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdownFn()

	tracer := otel.Tracer("client")

	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure(),
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

	meter := global.Meter("test-meter")

	// Recorder metric example
	counter := metric.Must(meter).
		NewInt64Counter(
			"an_important_metric",
			metric.WithDescription("Measures the cumulative epicness of the app"),
		)
	counter.Add(ctx, 1, commonLabels...)

	md := metadata.Pairs(
		"timestamp", time.Now().Format(time.StampNano),
		"client-id", "web-api-client-us-east-1",
		"user-id", "some-test-user-id",
	)
	ctx = metadata.NewOutgoingContext(ctx, md)

	req := &weatherpb.WeatherRequest{
		Location: "dublin",
	}

	res, err := c.GetCurrentWeather(ctx, req)
	if err != nil {
		span.RecordError(err)
		return
	}

	span.AddEvent("Response", trace.WithAttributes(
		attribute.String("condition", res.Condition),
		attribute.Float64("temperature", res.Temperature),
	))
}
