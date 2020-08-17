package main

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	apitrace "go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/stdout"
)

func main() {
	// first configure a TraceProvider to collect tracing information
	// A tracer is an object that tracks the currently active span and allows you to create (or activate) new spans. As spans are created and completed, the tracer dispatches them to an     exporter that can send the spans to a backend system for analysis
	// stdout which prints tracing information to the console
	exportOpts := []stdout.Option{
		stdout.WithQuantiles([]float64{0.5}),
		stdout.WithPrettyPrint(),
	}
	// Registers both a trace and meter Provider globally.
	pusher, err := stdout.InstallNewPipeline(exportOpts, nil)
	if err != nil {
		log.Fatal("Could not initialize stdout exporter:", err)
	}
	defer pusher.Stop()

	tracer := global.TraceProvider().Tracer("ex.com/basic")

	//  create a Span object. "run" is span's name.
	ctx, span := tracer.Start(context.Background(), "run")
	// Attributes allow you to add name/value pairs to describe the span
	span.SetAttributes(kv.String("platform", "osx"))
	span.SetAttributes(kv.String("version", "1.2.3"))
	// Events represent an event that occurred at a specific time within a span’s workload.
	span.AddEvent(ctx, "event in foo", kv.String("name", "foo1"))

	// set attributes for child span
	attributes := []kv.KeyValue{
		kv.String("platform", "osx"),
		kv.String("version", "1.2.3"),
	}

	// Add a child span to the existing span
	ctx, child := tracer.Start(ctx, "baz", apitrace.WithAttributes(attributes...))

	// end span
	child.End()
	span.End()

	// now parent span is "baz" in ctx
	_ = tracer.WithSpan(ctx, "foo1",
		func(ctx context.Context) error {
			tracer.WithSpan(ctx, "bar1",
				func(ctx context.Context) error {
					tracer.WithSpan(ctx, "baz1",
						func(ctx context.Context) error {
							return nil
						},
					)
					return nil
				},
			)
			return nil
		},
	)
}
