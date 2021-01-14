package main

import (
	"context"
	"example/pkg/pb"
	"fmt"

	"github.com/apache/dubbo-go/config"
)

func init() {
	config.SetProviderService(NewGreeterProvider())
}

type GreeterProvider struct {
	*pb.GreeterProviderBase
}

func NewGreeterProvider() *GreeterProvider {
	return &GreeterProvider{
		GreeterProviderBase: &pb.GreeterProviderBase{},
	}
}

func (g *GreeterProvider) SayHello(ctx context.Context, req *pb.HelloRequest) (reply *pb.HelloReply, err error) {
	fmt.Printf("req: %v", req)
	return &pb.HelloReply{Message: "this is message from reply"}, nil
}

func (g *GreeterProvider) Reference() string {
	return "GrpcGreeterImpl"
}
