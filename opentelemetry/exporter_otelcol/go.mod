module example

go 1.15

require (
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/otlp v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	google.golang.org/grpc v1.34.0
)

// v0.15.0 has panic("send on closed channel") when pusher.Stop()
replace (
	go.opentelemetry.io/otel => /home/chen/test/opentelemetry-go-master
	go.opentelemetry.io/otel/exporters/otlp => /home/chen/test/opentelemetry-go-master/exporters/otlp
	go.opentelemetry.io/otel/sdk => /home/chen/test/opentelemetry-go-master/sdk
)