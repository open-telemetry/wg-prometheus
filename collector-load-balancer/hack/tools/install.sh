#!/usr/bin/env sh

echo "installing goimports"

go install golang.org/x/tools/cmd/goimports

echo "installing protobuf and grpc gateway"

go install \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
  github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
  google.golang.org/protobuf/cmd/protoc-gen-go \
  google.golang.org/grpc/cmd/protoc-gen-go-grpc

echo "installing controller-gen and kustomize"

go install \
  sigs.k8s.io/controller-tools/cmd/controller-gen \
  sigs.k8s.io/kustomize/kustomize/v3
