package main

import (
	"context"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/zhipenwang/grpc_gateway/internal/conf"
	rpc_proto "github.com/zhipenwang/grpc_gateway/internal/grpcfile"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
)

type errBody struct {
	Code	int	   `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func main() {
	ctx := context.Background()
	dopts := []grpc.DialOption{grpc.WithInsecure()}
	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(CustomIncomingHeader),
		runtime.WithOutgoingHeaderMatcher(CustomOutcomingHeader),
		runtime.WithForwardResponseOption(CustomeHttpResponse),
		runtime.WithProtoErrorHandler(CustomHttpError),
	)
	if err := rpc_proto.RegisterHelloHandlerFromEndpoint(ctx, gwMux, conf.ServerAddr, dopts); err != nil {
		log.Fatalf("failed to register gw server, err=%v", err)
	}
	if err := http.ListenAndServe(conf.ServerHttpAddr, gwMux); err != nil {
		log.Fatalf("failed http server, err=%v", err)
	}
}

/**
自定义请求头的key，将http中的header—key转换为自定义的grpc-header-key，同时保留其他的映射规则
default中会把 runtime-context.go中isPermanentHTTPHeader，不符合的key去掉
如果是直接返回key的话，就不会被过滤，直接传输到grpc-header了

### 注意：key需要首字母大写；http层的header规则

example：
下面替换后，到grpc-header的结果是：
Origin-Key	=> origin-key
Custom-Key	=> grpc-metadata-custom-key
其他未指明替换的会走default进行替换，如：test-header-key，由于过滤，在grpc-header中就不会存在
 */
func CustomIncomingHeader(header_key string) (string, bool) {
	switch header_key {
	case "Origin-Key":
		return header_key, true
	case "Custom-Key":
		return "Grpc-Metadata-" + header_key, true
	default:
		return runtime.DefaultHeaderMatcher(header_key)
	}
}

/**
自定义响应头的key，将grpc中的header—key转换为自定义的http-header-key，同时保留其他的映射规则
default中会把 runtime-context.go中isPermanentHTTPHeader，不符合的key去掉
如果是直接返回key的话，就不会被过滤，直接传输到http-header了

### 注意：key按照grpc返回的响应header-key书写即可，http收到的header-key会是首字母大写的（http层的header规则）

example：
下面替换后，到http-header的结果是：
response-code	=> Response-Code
其他未指明替换的会走default进行替换，如：test-header-key，由于过滤，在http-header中就不会存在
*/
func CustomOutcomingHeader(header_key string) (string, bool) {
	switch header_key {
	case "response-code":
		return header_key, true
	default:
		return runtime.DefaultHeaderMatcher(header_key)
	}
}

/**
请求正常
更改响应body或者设置响应header
 */
func CustomeHttpResponse(ctx context.Context, w http.ResponseWriter, message proto.Message) error {
	// 设置响应header-key
	w.Header().Set("Response-Header-Code", "123456")

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		return nil
	}
	if vals := md.HeaderMD.Get("response-code"); len(vals) > 0 {
		// 设置header-http-code
		w.WriteHeader(http.StatusAccepted)
		// 删除不需要的grpc-header和http-header
		delete(md.HeaderMD, "response-code")
		delete(w.Header(), "Grpcgateway-Content-Type") // 发现删除不了，有点问题
	}

	// 目前发现是在响应的body之前追加了这个字符串
	w.Write([]byte("change response msg"))

	return nil
}

/**
一元rpc请求失败
自定义返回错误信息
 */
func CustomHttpError(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, req *http.Request, err error)  {
	const fallback = `{"code": 400, "message": "failed to marshal error message"}`
	w.Header().Set("Content-Type", marshaler.ContentType())
	if grpc_err, ok := status.FromError(err); ok {
		// 将grpc-code转为http-code
		http_code := runtime.HTTPStatusFromCode(grpc_err.Code())
		w.WriteHeader(http_code)
		// 定义响应body的参数，通过struct转json返回给调用方
		err_body := errBody{
			Code: http_code,
			Message: grpc_err.Message(),
		}
		jErr := json.NewEncoder(w).Encode(err_body)
		if jErr != nil {
			w.Write([]byte(fallback))
		}
	}
}