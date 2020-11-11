package main

import (
	"context"
	"github.com/zhipenwang/grpc_gateway/internal/conf"
	rpc_proto "github.com/zhipenwang/grpc_gateway/internal/grpcfile"
	"log"
	"net"
	"google.golang.org/grpc"
)

type HelloServer struct {

}

func (h *HelloServer) SayHello(ctx context.Context, request *rpc_proto.Request) (*rpc_proto.Response, error) {
	log.Printf("receive msg: %v", request)
	return &rpc_proto.Response{
		Message: "hello " + request.Name,
	}, nil
}

func main() {

	listen, err := net.Listen("tcp", conf.ServerAddr)
	if err != nil {
		log.Fatalf("failed to listen, err=%v", err)
	}
	grpcServer := grpc.NewServer()
	rpc_proto.RegisterHelloServer(grpcServer, &HelloServer{})
	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("grpc server failed, err=%v", err)
	}
}
