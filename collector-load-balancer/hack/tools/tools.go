// +build tools

package tools

// Based on https://github.com/go-modules-by-example/index/blob/master/010_tools/README.md
// gRPC gateway https://github.com/grpc-ecosystem/grpc-gateway#installation

import (
	_ "golang.org/x/tools/cmd/goimports"

	// k8s
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen"
	_ "sigs.k8s.io/kustomize/kustomize/v3"

	// grpc
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
