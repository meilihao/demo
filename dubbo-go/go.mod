module example

go 1.16

replace google.golang.org/grpc v1.35.0 => google.golang.org/grpc v1.26.0 // 没用最新: gomod使用时etcd和grpc版本冲突解决 from [etcd go.mod](https://github.com/etcd-io/etcd/blob/master/go.mod)

require (
	github.com/apache/dubbo-go v1.5.5
	github.com/golang/protobuf v1.4.3
	go.uber.org/multierr v1.6.0 // indirect
	google.golang.org/grpc v1.35.0
)
