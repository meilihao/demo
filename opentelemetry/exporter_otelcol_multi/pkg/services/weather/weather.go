package main

import (
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"distributed-tracing-otel/pkg/weatherpb"

	"github.com/meilihao/golib/v2"
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
			attribute.String("metadata", golib.DumpToJsonString(md)),
		))
	}

	l, ok := s.locations[in.Location]
	if !ok {
		err := status.Error(codes.NotFound, "Location not found")
		span.RecordError(err, trace.WithAttributes(
			attribute.String("status", codes.NotFound.String()),
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

	shutdownFn, err := golib.SetupOTelSDK(ctx, otelEndpoint, "weather")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdownFn(ctx)

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

	go everLoop(ctx, s)
	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	log.Println("app exit")
}

func everLoop(ctx context.Context, s *grpc.Server) {
LOOP:
	for {
		select {
		case <-ctx.Done():
			// 收到信号退出无限循环
			log.Println("everLoop: 收到信号退出无限循环")
			s.Stop()
			break LOOP
		default:
			// 用一个 sleep 模拟业务逻辑
			time.Sleep(time.Second * 10)
		}
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
