package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"distributed-tracing-otel/pkg/lib"
	"distributed-tracing-otel/pkg/weatherpb"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type server struct {
	locations map[string]string
}

func (s *server) GetCurrentWeather(ctx context.Context, in *weatherpb.WeatherRequest) (*weatherpb.WeatherResponse, error) {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	// see https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		span.AddEvent("metadata", trace.WithAttributes(
			attribute.Any("metadata", md),
		))
	}

	l, ok := s.locations[in.Location]
	if !ok {
		err := status.Error(codes.NotFound, "Location not found")
		span.RecordError(err, trace.WithAttributes(
			attribute.Any("status", codes.NotFound),
		))
		return nil, err
	}

	span.AddEvent("Selected condition", trace.WithAttributes(
		attribute.String("condition", l),
		attribute.String("location", in.Location),
	))

	t, err := getTemperature(ctx)

	if err != nil {
		err := status.Error(codes.Unknown, err.Error())
		span.RecordError(err)
		return nil, err
	}

	span.AddEvent("Temperature received", trace.WithAttributes(
		attribute.Float64("temperature", t),
	))

	return &weatherpb.WeatherResponse{
		Condition:   l,
		Temperature: t,
	}, nil
}

var (
	tracer trace.Tracer
)

func main() {
	shutdownFn, err := lib.InitOTEL("openhello.net:55680", "weather")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdownFn()

	tracer = otel.Tracer("weather-tracer")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := &server{
		locations: map[string]string{
			"dublin":   "rainy",
			"galway":   "sunny",
			"limerick": "cloudy",
		},
	}

	s := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	weatherpb.RegisterWeatherServiceServer(s, server)

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// --- temperature
// WeatherForecast represents the response from temperature service
type WeatherForecast struct {
	TemperatureC int `json:"temperatureC"`
}

func getTemperature(ctx context.Context) (float64, error) {
	ctx, span := tracer.Start(ctx, "getTemperature")
	defer span.End()

	// client := &http.Client{
	// 	Transport: otelhttp.NewTransport(http.DefaultTransport),
	// } // 也可以, 但时序会错乱: do request before http receive
	client := http.DefaultClient

	req, _ := http.NewRequest("GET", "http://localhost:5000/WeatherForecast", nil)

	needHttpDetail := true
	if needHttpDetail {
		ctx, req = otelhttptrace.W3C(ctx, req)
		otelhttptrace.Inject(ctx, req)
	} else {
		req = req.WithContext(ctx)
		otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
	}

	res, err := client.Do(req)
	if err != nil {
		return 0.0, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0.0, err
	}
	defer res.Body.Close()

	span.AddEvent("http req done")

	wf, err := parseTemperatureResponse(body)
	return float64(wf.TemperatureC), err
}

func parseTemperatureResponse(body []byte) (WeatherForecast, error) {
	wf := &WeatherForecast{}
	err := json.Unmarshal(body, wf)
	return *wf, err
}
