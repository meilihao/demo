// from https://github.com/open-telemetry/opentelemetry-go/blob/main/example/otel-collector/main.go
// see [opentelemetry-java/QUICKSTART.md](https://github.com/open-telemetry/opentelemetry-java/blob/master/QUICKSTART.md)
// [Documentation / Go / Getting Started](https://opentelemetry.io/docs/go/getting-started/)
package lib

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitOTEL(endpoint, serviceName string) (func(), error) {
	if endpoint == "" {
		return func() {}, nil
	}

	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
		),
		resource.WithProcess(), // show span Process in jaeger
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create resource")
	}

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns

	// conn, err := grpc.DialContext(cctx, endpoint, grpc.WithInsecure()) // grpc.WithInsecure() Deprecated: use insecure.NewCredentials() instead.
	dialTimeout := time.Second * 10
	conn, err := grpc.DialContext(ctx, endpoint,
		grpc.WithTimeout(dialTimeout),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create gRPC connection to collector")
	}

	client := otlpmetricgrpc.NewClient(
		// otlpmetricgrpc.WithEndpoint(endpoint),
		// otlpmetricgrpc.WithReconnectionPeriod(50*time.Millisecond),
		// otlpmetricgrpc.WithTimeout(dialTimeout),
		// otlpmetricgrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()),
		otlpmetricgrpc.WithGRPCConn(conn),
	)

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create trace exporter")
	}

	logger, _ := zap.NewProduction()

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSpanProcessor(NewLogSpanProcessor(logger)),
	)

	exp, err := otlpmetric.New(ctx, client)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create metric exporter")
	}

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			exp,
		),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(2*time.Second),
	)

	// set global TracerProvider (the default is noopTracerProvider).
	otel.SetTracerProvider(tracerProvider)
	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	global.SetMeterProvider(pusher)

	if err = pusher.Start(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to start metric controller")
	}

	logger.Info("init InitOTEL done")
	return func() {
		// Shutdown will flush any remaining spans.
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.Error(err.Error(), zap.String("reason", "failed to shutdown provider"))
		}

		// Push any last metric events to the exporter.
		if err := pusher.Stop(context.Background()); err != nil {
			logger.Error(err.Error(), zap.String("reason", "failed to stop exporter"))
		}
	}, nil
}
