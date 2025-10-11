package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"

	pb "gw/github.com/yourorg/yourprotos/gen/go/your/service/v1" // Update
)

var (
	// command-line options:
	// gRPC server endpoint
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:9090", "gRPC server endpoint")
)

type HTTPService struct {
	pb.UnimplementedYourServiceServer
}

var httpService = &HTTPService{}

func (h HTTPService) Echo(ctx context.Context, in *pb.StringMessage) (*pb.StringMessage, error) {
	grpclog.Infoln("收到请求: Echo")
	resp := new(pb.StringMessage)
	resp.Value = "get msg:" + in.Value
	return resp, nil
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcServer := grpc.NewServer()
	pb.RegisterYourServiceServer(grpcServer, httpService)
	go func() {
		listen, err := net.Listen("tcp", *grpcServerEndpoint)
		if err != nil {
			grpclog.Fatalf("failed to listen: %v", err)
		}

		if err := grpcServer.Serve(listen); err != nil {
			grpclog.Fatalf("failed to serve: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	// RegisterYourServiceHandlerFromEndpoint创建的ClientConn只有请求grpc server时才会建立连接
	err := pb.RegisterYourServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(":8081", mux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		grpclog.Fatal(err)
	}
}
