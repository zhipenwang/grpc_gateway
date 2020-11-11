# grpc_gateway
http转grpc


### 下载gateway生成工具
```
https://github.com/grpc-ecosystem/grpc-gateway

package tools

import (
    _ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
    _ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
    _ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
    _ "google.golang.org/protobuf/cmd/protoc-gen-go"
)

Run go mod tidy

go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

### 编写proto文件支持api
```
import "google/api/annotations.proto";
rpc SayHello (Request) returns (Response) {
    // 引入google api实现grpc转http
    // 定义http请求路由
    option (google.api.http) = {
      post: "/http/hello"
      body: "*"
    };
}

第三方文件所在：https://github.com/grpc-ecosystem/grpc-gateway/tree/master/third_party/googleapis/google/api
（可以在你本地mod的缓存库里找）
需要把编译需要的第三方文件：annotations.proto、http.proto
复制到你的proto文件目录下，本地导入

然后执行编译脚本：script/gen-proto.sh
```

#### 第一步生成未带http.api的proto，然后proto文件加上http.api，再生成gateway的proto
```
第一步
rpc SayHello (Request) returns (Response) {}
生成：protoc --proto_path=protos --go_out=plugins=grpc:internal/grpcfile hello.proto 

第二步：修改proto文件
rpc SayHello (Request) returns (Response) {
    // 引入google api实现grpc转http
    // 定义http请求路由
    option (google.api.http) = {
      post: "/http/hello"
      body: "*"
    };
}
生成：protoc -I . --proto_path=protos --grpc-gateway_out ./internal/grpcfile --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative  hello.proto
```

### 开发
#### grpc服务跟http服务分开部署
```
server_grpc: grpc服务
server_http: http服务
注：两个服务可独立部署，建议服务项目拆分为两个，互不影响

client: 客户端grpc请求

http请求：curl -XPOST http://127.0.0.1:8898/http/hello -d '{"name":"http test"}
```

#### grpc跟http部署到统一个服务，公用同一个端口
```
net/http中对http2的支持要求开启https，所以这里要求使用https服务
第一步：
先进行自签名证书申请(参考：https://www.bookstack.cn/read/topgoer/175655e5e1d49d52.md)
安装openssl
制作私钥（.key）
openssl genrsa -out server.key 2048
openssl ecparam -genkey -name secp384r1 -out server.key

自签名公钥(x509) (PEM-encodings .pem|.crt)
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
【其中common name-服务名称可以自定义，仅供测试哈，测试中我用grpc_gateway】

server_grpc_http: grpc与http服务
client_grpc_http: 客户端grpc请求(TLS请求)

http请求：curl -XPOST -k https://127.0.0.1:8899/http/hello -d '{"name":"123"}'
因为是自签名证书，需要curl中加 -k 参数
    -k, --insecure      允许连接到 SSL 站点，而不使用证书 (H)
```