package main

import (
	"context"
	"crypto/tls"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/zhipenwang/grpc_gateway/internal/conf"
	rpc_proto "github.com/zhipenwang/grpc_gateway/internal/grpcfile"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
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

	// TLS自签名证书
	creds, _ := credentials.NewServerTLSFromFile("../../internal/cert/server.pem", "../../internal/cert/server.key")
	grpcServer := grpc.NewServer(grpc.Creds(creds))
	rpc_proto.RegisterHelloServer(grpcServer, &HelloServer{})

	ctx := context.Background()
	// TLS自签名证书
	dcreds, _ := credentials.NewClientTLSFromFile("../../internal/cert/server.pem", "grpc_gateway")
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}
	gwMux := runtime.NewServeMux()
	if err := rpc_proto.RegisterHelloHandlerFromEndpoint(ctx, gwMux, conf.ServerAddr, dopts); err != nil {
		log.Fatalf("failed to register gw server, err=%v", err)
	}

	// http server
	mux := http.NewServeMux()
	mux.Handle("/", gwMux)

	srv := &http.Server{
		Addr: conf.ServerAddr,
		Handler: grpcHandlerFunc(grpcServer, mux),
		TLSConfig: getTLSConfig(),
	}

	if err := srv.Serve(tls.NewListener(listen, srv.TLSConfig)); err != nil {
		log.Fatalf("grpc & http server failed, err=%v", err)
	}
}

func getTLSConfig() *tls.Config {
	cert, _ := ioutil.ReadFile("../../internal/cert/server.pem")
	key, _ := ioutil.ReadFile("../../internal/cert/server.key")
	var demoKeyPair *tls.Certificate
	pair, err := tls.X509KeyPair(cert, key)
	if err != nil {
		grpclog.Fatalf("TLS KeyPair err: %v\n", err)
	}
	demoKeyPair = &pair
	return &tls.Config{
		Certificates: []tls.Certificate{*demoKeyPair},
		NextProtos:   []string{http2.NextProtoTLS}, // HTTP2 TLS支持
	}
}

// grpcHandlerFunc returns an http.Handler that delegates to grpcServer on incoming gRPC
// connections or otherHandler otherwise. Copied from cockroachdb.
func grpcHandlerFunc(grpcServer *grpc.Server, otherHandler http.Handler) http.Handler {
	if otherHandler == nil {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			grpcServer.ServeHTTP(w, r)
		})
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			otherHandler.ServeHTTP(w, r)
		}
	})
}