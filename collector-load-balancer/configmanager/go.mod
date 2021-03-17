module github.com/aws-observability/collector-load-balancer/configmanager

go 1.16

require (
	github.com/aws-observability/collector-load-balancer/proto/generated/clb v0.0.0-00010101000000-000000000000
	github.com/dyweb/gommon v0.0.13
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.3.0
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.16.0
	google.golang.org/grpc v1.36.0
)

replace github.com/aws-observability/collector-load-balancer/proto/generated/clb => ../proto/generated/clb
