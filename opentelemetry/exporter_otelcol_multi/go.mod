module distributed-tracing-otel

go 1.15

require (
	github.com/gin-gonic/gin v1.6.3
	github.com/golang/protobuf v1.4.3
	github.com/pkg/errors v0.9.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.15.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.15.1
	go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace v0.15.1
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.15.1
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/otlp v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

// v0.15.0 has panic("send on closed channel") when pusher.Stop()
replace (
	go.opentelemetry.io/otel => /home/chen/test/opentelemetry-go-master
	go.opentelemetry.io/otel/exporters/otlp => /home/chen/test/opentelemetry-go-master/exporters/otlp
	go.opentelemetry.io/otel/sdk => /home/chen/test/opentelemetry-go-master/sdk
)
