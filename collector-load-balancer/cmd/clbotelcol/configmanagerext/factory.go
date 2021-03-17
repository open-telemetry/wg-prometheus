package configmanagerext

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/extension/extensionhelper"

	"github.com/aws-observability/collector-load-balancer/configmanager"
)

const (
	typeStr configmodels.Type = "clb_config_manager"
)

func NewFactory() component.ExtensionFactory {
	return extensionhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		createExtension,
	)
}

func createDefaultConfig() configmodels.Extension {
	return &Config{
		ExtensionSettings: configmodels.ExtensionSettings{
			TypeVal: typeStr,
			NameVal: string(typeStr),
		},
		ServerConfig: configmanager.DefaultServerConfig(),
	}
}

func createExtension(ctx context.Context, params component.ExtensionCreateParams, cfg configmodels.Extension) (component.Extension, error) {
	config := cfg.(*Config)
	srv, err := configmanager.NewServer(config.ServerConfig, configmanager.ServerOptions{Logger: params.Logger})
	if err != nil {
		return nil, err
	}
	return &extension{srv: srv}, nil
}
