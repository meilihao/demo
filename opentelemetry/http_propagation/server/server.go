package main

import (
	"io"
	"log"
	"net/http"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
)

func helloHandler(w http.ResponseWriter, req *http.Request) {
	tracer := global.TraceProvider().Tracer("example/server")

	// Extracts the conventional HTTP span attributes,
	// distributed context tags, and a span context for
	// tracing this request.
	attrs, entries, spanCtx := httptrace.Extract(req.Context(), req)
	ctx := req.Context()
	if spanCtx.IsValid() {
		ctx = trace.ContextWithRemoteSpanContext(ctx, spanCtx)
	}

	// Apply the correlation context tags to the request
	// context.
	req = req.WithContext(correlation.ContextWithMap(ctx, correlation.NewMap(correlation.MapUpdate{
		MultiKV: entries,
	})))

	// Start the server-side span, passing the remote
	// child span context explicitly.
	_, span := tracer.Start(
		req.Context(),
		"hello",
		trace.WithAttributes(attrs...),
	)
	defer span.End()

	io.WriteString(w, "Hello, world!\n")
}

func main() {
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

	http.HandleFunc("/hello", helloHandler)
	err = http.ListenAndServe(":7777", nil)
	if err != nil {
		panic(err)
	}
}
