# Multi Services
- [example from "Distributed Tracing with OpenTelemetry — Part 2](https://levelup.gitconnected.com/distributed-tracing-with-opentelemetry-part-2-cc5a9a8aa88c)

This sample application shows how to implement distributed trancing using OpenTelemetry in grpc and http by golang.

![arch](/opentelemetry/exporter_otelcol_multi/misc/1nCd2RjWGBqrWj7HEiKkosQ.png)

This case is expand from [exporter_otelcol](/opentelemetry/exporter_otelcol) and [other examples](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation)

> exporter is OpenTelemetry Collector.

## gen proto
```bash
# cd pkg/weatherpb
# protoc --go_out=plugins=grpc:. *.proto
```

## run
### temperature
```bash
cd pkg/services/temperature
go build
./temperature
```

### weather

```bash
cd pkg/services/weather
go build
./weather
```

### client

```bash
cd pkg/services/client
go build
./client
```
