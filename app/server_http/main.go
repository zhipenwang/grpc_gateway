package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/zhipenwang/grpc_gateway/internal/conf"
	rpc_proto "github.com/zhipenwang/grpc_gateway/internal/grpcfile"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func main() {
	ctx := context.Background()
	dopts := []grpc.DialOption{grpc.WithInsecure()}
	gwMux := runtime.NewServeMux()
	if err := rpc_proto.RegisterHelloHandlerFromEndpoint(ctx, gwMux, conf.ServerAddr, dopts); err != nil {
		log.Fatalf("failed to register gw server, err=%v", err)
	}
	if err := http.ListenAndServe(conf.ServerHttpAddr, gwMux); err != nil {
		log.Fatalf("failed http server, err=%v", err)
	}
}