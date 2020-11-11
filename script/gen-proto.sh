#!/usr/bin/env bash

#生成go proto文件
protoc --proto_path=protos --go_out=plugins=grpc:internal/grpcfile hello.proto

#生成gateway文件
protoc -I . --proto_path=protos --grpc-gateway_out ./internal/grpcfile --grpc-gateway_opt logtostderr=true --grpc-gateway_opt paths=source_relative  hello.proto