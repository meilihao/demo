// from https://github.com/open-telemetry/opentelemetry-go/blob/master/example/otel-collector/main.go
// see [opentelemetry-java/QUICKSTART.md](https://github.com/open-telemetry/opentelemetry-java/blob/master/QUICKSTART.md)
// [Documentation / Go / Getting Started](https://opentelemetry.io/docs/go/getting-started/)
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

var (
	// for no trace and no metric
	enableTelemetry = true
)

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() func() {
	ctx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint("openhello.net:55680"),
		otlpgrpc.WithDialOption(grpc.WithBlock()), // useful for testing
	)
	exp, err := otlp.NewExporter(ctx, driver)
	handleErr(err, "failed to create exporter")

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String("test-service"),
		),
	)
	handleErr(err, "failed to create resource")

	logger, _ := zap.NewProduction()
	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithSpanProcessor(NewLogSpanProcessor(logger)),
	)

	cont := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exp,
		),
		controller.WithPusher(exp),
		controller.WithCollectPeriod(2*time.Second),
	)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})
	// set global TracerProvider (the default is noopTracerProvider).
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(cont.MeterProvider())
	handleErr(cont.Start(context.Background()), "failed to start controller")

	return func() {
		// Shutdown will flush any remaining spans.
		handleErr(tracerProvider.Shutdown(ctx), "failed to shutdown TracerProvider")

		// Push any last metric events to the exporter.
		handleErr(cont.Stop(context.Background()), "failed to stop controller")
	}
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	log.Printf("Waiting for connection...")

	if enableTelemetry {
		shutdown := initProvider()
		defer shutdown()
	}

	log.Println("provider init done")

	tracer := otel.Tracer("test-tracer")
	meter := otel.Meter("test-meter")

	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	// <namespace>_an_important_metric{labelA="chocolate",labelB="raspberry",labelC="vanilla"} 2
	commonLabels := []label.KeyValue{
		label.String("labelA", "chocolate"),
		label.String("labelB", "raspberry"),
		label.String("labelC", "vanilla"),
	}

	// Recorder metric example
	valuerecorder := metric.Must(meter).
		NewFloat64Counter(
			"an_important_metric",
			metric.WithDescription("Measures the cumulative epicness of the app"),
		).Bind(commonLabels...)
	defer valuerecorder.Unbind()

	// work begins
	ctx, span := tracer.Start(
		context.Background(),
		"CollectorExporter-Example",
		trace.WithAttributes(commonLabels...))
	defer span.End()
	for i := 0; i < 2; i++ {
		_, iSpan := tracer.Start(ctx, fmt.Sprintf("Sample-%d", i))
		log.Printf("Doing really hard work (%d / 10)\n", i+1)
		valuerecorder.Add(ctx, 1.0)

		iSpan.SetStatus(codes.Error, "error")
		iSpan.SetAttributes(label.Bool("is_done", true))

		// equal opentracing's span.LogFields
		iSpan.AddEvent("failed", trace.WithAttributes(label.String("reason", "test")))

		<-time.After(time.Second)
		iSpan.End()
	}

	log.Printf("Done!")
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
