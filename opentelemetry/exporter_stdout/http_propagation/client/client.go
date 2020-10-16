package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
)

func main() {
	exportOpts := []stdout.Option{
		stdout.WithQuantiles([]float64{0.5}),
		stdout.WithPrettyPrint(),
	}

	pusher, err := stdout.InstallNewPipeline(exportOpts, nil)
	if err != nil {
		log.Fatal("Could not initialize stdout exporter:", err)
	}
	defer pusher.Stop()

	client := http.DefaultClient
	ctx := correlation.NewContext(context.Background(),
		kv.String("username", "donuts"),
	)

	var body []byte

	tracer := global.TraceProvider().Tracer("example/client")
	err = tracer.WithSpan(ctx, "client hello",
		func(ctx context.Context) error {
			req, _ := http.NewRequest("GET", "http://localhost:7777/hello", nil)

			ctx, req = httptrace.W3C(ctx, req)
			httptrace.Inject(ctx, req)
			res, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			body, err = ioutil.ReadAll(res.Body)
			_ = res.Body.Close()
			trace.SpanFromContext(ctx).SetStatus(200, "OK")
			return err
		})

	if err != nil {
		panic(err)
	}
	time.Sleep(time.Second * 5)

}
