module github.com/aws-observability/collector-load-balancer/cmd/clbotelcol

go 1.16

require (
	github.com/aws-observability/collector-load-balancer/configmanager v0.0.0-00010101000000-000000000000
	github.com/dyweb/gommon v0.0.13
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsemfexporter v0.22.0
	github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsprometheusremotewriteexporter v0.22.0
	github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor v0.22.0
	github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver v0.22.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	go.opentelemetry.io/collector v0.22.0
)

replace github.com/aws-observability/collector-load-balancer/configmanager => ../../configmanager

replace github.com/aws-observability/collector-load-balancer/proto/generated/clb => ../../proto/generated/clb
