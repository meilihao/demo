// [opentelemetry-go api status](https://opentelemetry.io/docs/go/) : go sdk未实现log api, 自己动手
package main

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/label"
	exporttrace "go.opentelemetry.io/otel/sdk/export/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	itrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// LogSpanProcessor is a SpanProcessor that customized export logs
type LogSpanProcessor struct {
	logger   *zap.Logger
	stopOnce sync.Once
}

var _ sdktrace.SpanProcessor = (*LogSpanProcessor)(nil)

// NewLogSpanProcessor creates a new LogSpanProcessor that will customized export logs.
func NewLogSpanProcessor(logger *zap.Logger) *LogSpanProcessor {
	lsp := &LogSpanProcessor{
		logger: logger,
	}

	return lsp
}

// OnStart method does nothing.
func (lsp *LogSpanProcessor) OnStart(parent context.Context, s sdktrace.ReadWriteSpan) {}

// OnEnd method enqueues a ReadOnlySpan for later processing.
func (lsp *LogSpanProcessor) OnEnd(s sdktrace.ReadOnlySpan) {
	// Do not enqueue spans if we are just going to drop them.
	if lsp.logger == nil {
		return
	}

	for _, e := range s.Events() {
		lsp.logger.Info(e.Name, exportKVs(s.SpanContext(), e)...)
	}
}

func exportKVs(sctx itrace.SpanContext, e exporttrace.Event) []zap.Field {
	ls := make([]zap.Field, 0, len(e.Attributes)+3)
	// TODO replace default ts
	ls = append(ls, zap.Time("time", e.Time))
	ls = append(ls, zap.String("trace_id", sctx.TraceID.String()))
	ls = append(ls, zap.String("span_id", sctx.SpanID.String()))

	for _, attr := range e.Attributes {
		// TODO add others type
		switch attr.Value.Type() {
		case label.STRING:
			ls = append(ls, zap.String(string(attr.Key), attr.Value.AsString()))
		default:
			ls = append(ls, zap.Any(string(attr.Key), attr.Value))
		}
	}

	return ls
}

// Shutdown flushes the queue and waits until all spans are processed.
// It only executes once. Subsequent call does nothing.
func (lsp *LogSpanProcessor) Shutdown(ctx context.Context) error {
	var err error
	lsp.stopOnce.Do(func() {
		err = lsp.logger.Sync()
	})
	return err
}

// ForceFlush exports all ended spans that have not yet been exported.
func (lsp *LogSpanProcessor) ForceFlush() {
}
