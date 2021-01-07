// from https://github.com/open-telemetry/opentelemetry-go/blob/master/example/otel-collector/main.go
// see [opentelemetry-java/QUICKSTART.md](https://github.com/open-telemetry/opentelemetry-java/blob/master/QUICKSTART.md)
// [Documentation / Go / Getting Started](https://opentelemetry.io/docs/go/getting-started/)
package lib

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func InitOTEL(endpoint, serviceName string) (func(), error) {
	if endpoint == "" {
		return func() {}, nil
	}

	ctx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	driver := otlp.NewGRPCDriver(
		otlp.WithInsecure(),
		otlp.WithEndpoint(endpoint),
		otlp.WithGRPCDialOption(grpc.WithBlock()), // useful for testing
	)
	exp, err := otlp.NewExporter(ctx, driver)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create exporter")
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create resource")
	}

	logger, _ := zap.NewProduction()
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSpanProcessor(NewLogSpanProcessor(logger)),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	// set global TracerProvider (the default is noopTracerProvider).
	otel.SetTracerProvider(tracerProvider)

	pusher := push.New(
		basic.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		exp,
		push.WithPeriod(2*time.Second),
	)

	otel.SetMeterProvider(pusher.MeterProvider())
	pusher.Start()

	return func() {
		// Shutdown will flush any remaining spans.
		if err := tracerProvider.Shutdown(ctx); err != nil {
			logger.Error(err.Error(), zap.String("reason", "failed to shutdown provider"))
		}

		// Push any last metric events to the exporter.
		pusher.Stop()
		if err := exp.Shutdown(ctx); err != nil {
			logger.Error(err.Error(), zap.String("reason", "failed to stop exporter"))
		}
	}, nil
}
