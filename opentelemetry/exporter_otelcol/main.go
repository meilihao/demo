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
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
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
	logger          *zap.Logger
)

type loggerKey struct{}

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

	atom := zap.NewAtomicLevelAt(zap.DebugLevel)
	logger, _ = zap.Config{
		Level:       atom,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}.Build(zap.AddCallerSkip(1))

	bsp := sdktrace.NewBatchSpanProcessor(exp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
		//sdktrace.WithSpanProcessor(NewLogSpanProcessor(logger)), //废弃, 使用SpanLog代替
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
	global.SetMeterProvider(cont.MeterProvider())
	handleErr(cont.Start(context.Background()), "failed to start controller")

	return func() {
		// Shutdown will flush any remaining spans and shut down the exporter.
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
	meter := global.Meter("test-meter")

	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	// <namespace>_an_important_metric{labelA="chocolate",labelB="raspberry",labelC="vanilla"} 2
	commonLabels := []attribute.KeyValue{
		attribute.String("labelA", "chocolate"),
		attribute.String("labelB", "raspberry"),
		attribute.String("labelC", "vanilla"),
	}

	// Recorder metric example
	valuerecorder := metric.Must(meter).
		NewFloat64Counter(
			"an_important_metric",
			metric.WithDescription("Measures the cumulative epicness of the app"),
		).Bind(commonLabels...)
	defer valuerecorder.Unbind()

	// work begins
	baseCtx := context.WithValue(context.Background(), loggerKey{}, logger)
	ctx, span := tracer.Start(
		baseCtx,
		"CollectorExporter-Example",
		trace.WithAttributes(commonLabels...))
	defer span.End()
	for i := 0; i < 2; i++ {
		iCtx, iSpan := tracer.Start(ctx, fmt.Sprintf("Sample-%d", i))
		log.Printf("Doing really hard work (%d / 10)\n", i+1)
		valuerecorder.Add(ctx, 1.0)

		iSpan.SetStatus(codes.Error, "error")
		iSpan.SetAttributes(attribute.Bool("is_done", true))

		// equal opentracing's span.LogFields
		iSpan.AddEvent("failed", trace.WithAttributes(attribute.String("reason", "test")))
		SpanLog(iCtx, iSpan, zap.DebugLevel, "debug log", attribute.String("reason", "test"))
		SpanLog(iCtx, iSpan, zap.InfoLevel, "info log", attribute.String("reason", "test"))
		SpanLog(iCtx, iSpan, zap.InfoLevel, "test")

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

func SpanLog(ctx context.Context, span trace.Span, l zapcore.Level, msg string, kv ...attribute.KeyValue) {
	var logger *zap.Logger
	if tmp := ctx.Value(loggerKey{}); tmp == nil {
		return
	} else {
		logger = tmp.(*zap.Logger)
	}

	if ce := logger.Check(l, msg); ce != nil {
		sctx := span.SpanContext()

		fs := make([]zap.Field, 0, len(kv)+2)
		fs = append(fs, zap.String("trace_id", sctx.TraceID.String()))
		fs = append(fs, zap.String("span_id", sctx.SpanID.String()))

		if len(kv) > 0 {
			for _, attr := range kv {
				switch attr.Value.Type() {
				case attribute.STRING:
					fs = append(fs, zap.String(string(attr.Key), attr.Value.AsString()))
				default:
					fs = append(fs, zap.Any(string(attr.Key), attr.Value))
				}
			}
		}

		ce.Write(fs...)

		kv = append(kv, attribute.String("level", l.String()))
		span.AddEvent(msg, trace.WithAttributes(kv...))
	}
}
