module github.com/aws-observability/collector-load-balancer/hack/tools

go 1.16

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.3.0
	golang.org/x/tools v0.1.0
	google.golang.org/grpc/cmd/protoc-gen-go-grpc v1.1.0
	google.golang.org/protobuf v1.25.1-0.20201208041424-160c7477e0e8
	sigs.k8s.io/controller-tools v0.5.0
	sigs.k8s.io/kustomize/kustomize/v3 v3.10.0
)
