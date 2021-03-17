package configmanagerext

import (
	"go.opentelemetry.io/collector/config/configmodels"

	"github.com/aws-observability/collector-load-balancer/configmanager"
)

type Config struct {
	configmodels.ExtensionSettings `mapstructure:",squash"`
	configmanager.ServerConfig     `mapstructure:",squash"`
}
