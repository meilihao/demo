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
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	shutdownFn, err := lib.InitOTEL("openhello.net:55680", "client")
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
	ctx, span := tracer.Start(context.Background(), "GetCurrentWeather")
	defer span.End()

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
