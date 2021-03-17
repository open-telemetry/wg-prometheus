module github.com/aws-observability/collector-load-balancer

go 1.15

require (
	github.com/aws-observability/collector-load-balancer/configmanager v0.0.0-00010101000000-000000000000
	github.com/aws-observability/collector-load-balancer/proto/generated/clb v0.0.0-00010101000000-000000000000
	github.com/dyweb/gommon v0.0.13
	github.com/go-kit/kit v0.10.0
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.2.0
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/prometheus/common v0.17.0
	github.com/prometheus/prometheus v1.8.2-0.20210220213500-8c8de46003d1
	github.com/stretchr/testify v1.7.0
	go.uber.org/zap v1.16.0
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a
	google.golang.org/grpc v1.36.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3

)

replace github.com/aws-observability/collector-load-balancer/configmanager => ./configmanager

replace github.com/aws-observability/collector-load-balancer/proto/generated/clb => ./proto/generated/clb
