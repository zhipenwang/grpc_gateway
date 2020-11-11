package main

import (
	"context"
	"github.com/zhipenwang/grpc_gateway/internal/conf"
	rpc_proto "github.com/zhipenwang/grpc_gateway/internal/grpcfile"
	"google.golang.org/grpc"
	"log"
)

func main() {
	conn, err := grpc.Dial(conf.ServerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("conn failed, err=%v", err)
	}
	defer conn.Close()

	client := rpc_proto.NewHelloClient(conn)
	res, err := client.SayHello(context.Background(), &rpc_proto.Request{
		Name: "grpc client",
	})
	if err != nil {
		log.Printf("send err=%v", err)
	}
	log.Printf("response = %v", res)
}