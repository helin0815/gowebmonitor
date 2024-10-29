package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"examples/grpcapp/api/hello/v1"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type helloService struct {
	hello.UnsafeHelloServiceServer
}

func (h helloService) SayHello(ctx context.Context, request *hello.SayHelloRequest) (*hello.SayHelloResponse, error) {
	return &hello.SayHelloResponse{Message: fmt.Sprintf("Hello, %s!", request.Name)}, nil
}

func main() {
	lse := lsego.New()
	lse.SetupSwagger("/openapiv2/", os.DirFS("./api/openapiv2/"))
	lse.GrpcRegister(func(s *grpc.Server, gwMux *runtime.ServeMux) {
		// Register your GRPC server here.
		hello.RegisterHelloServiceServer(s, &helloService{})

		// Register your HTTP server here.
		ctx := context.Background()
		_ = hello.RegisterHelloServiceHandlerServer(ctx, gwMux, &helloService{})
	})
	lse.RegisterOnShutdown(func() {
		// shutdown something

		// sleep 10s to simulate some shutdown logic
		time.Sleep(time.Second * 10)
	})
	// only register the grpc server, not contain the http server
	// lse.GRPCServe(":9001")

	// register the grpc server and grpc gateway http server
	lse.GRPCServeWithGateway(":9001", ":9002")
}
