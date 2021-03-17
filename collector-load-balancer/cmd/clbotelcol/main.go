// clbotelcol is Collector Load Balancer's distribution of Open Telemetry Collector.
// It comes with ConfigManager as otel extension.
package main

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/dyweb/gommon/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/service"
	"go.opentelemetry.io/collector/service/defaultcomponents"

	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsemfexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/exporter/awsprometheusremotewriteexporter"
	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/metricstransformprocessor"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/awsecscontainermetricsreceiver"

	"github.com/aws-observability/collector-load-balancer/cmd/clbotelcol/configmanagerext"
)

func main() {
	if err := runCollector(); err != nil {
		log.Fatal(err)
	}
}

const (
	envKey = "AOT_CONFIG_CONTENT"
)

func runCollector() error {
	factories, err := components()
	if err != nil {
		return errors.Wrap(err, "error register components")
	}
	info := component.ApplicationStartInfo{
		ExeName:  "clbotelcol",
		LongName: "Collector Load Balancer OTEL Collector",
		Version:  "0.0.1",
		GitHash:  "this is a sha1 value",
	}
	return runInteractive(service.Parameters{
		ApplicationStartInfo: info,
		Factories:            factories,
		ConfigFactory:        loadConfig(),
	})
}

func runInteractive(params service.Parameters) error {
	app, err := service.New(params)
	if err != nil {
		return fmt.Errorf("failed to construct the application: %w", err)
	}

	err = app.Run()
	if err != nil {
		return fmt.Errorf("application run finished with error: %w", err)
	}

	return nil
}

// ----------------------------------------------------------------------------
// Copy Start
// TODO(pingleig() copied from aoc pkg/config/config_factory.go

func loadConfig() func(otelViper *viper.Viper, cmd *cobra.Command, f component.Factories) (*configmodels.Config, error) {
	return func(otelViper *viper.Viper, cmd *cobra.Command, f component.Factories) (*configmodels.Config, error) {
		// aws-otel-collector supports loading yaml config from Env Var
		// including SSM parameter store for ECS use case
		if configContent, ok := os.LookupEnv(envKey); ok {
			log.Printf("Reading AOT config from from environment: \n%v\n", configContent)
			return readConfigString(otelViper, f, configContent)
		}

		// use OTel yaml config from input
		return service.FileLoaderConfigFactory(otelViper, cmd, f)
	}
}

// readConfigString set aws-otel-collector config from env var
func readConfigString(v *viper.Viper,
	factories component.Factories,
	configContent string) (*configmodels.Config, error) {
	v.SetConfigType("yaml")
	var configBytes = []byte(configContent)
	err := v.ReadConfig(bytes.NewBuffer(configBytes))
	if err != nil {
		return nil, fmt.Errorf("error loading config %v", err)
	}
	return config.Load(v, factories)
}

// Copy End
// ----------------------------------------------------------------------------

func components() (component.Factories, error) {
	merr := errors.NewMultiErr()

	factories, err := defaultcomponents.Components()
	if err != nil {
		return component.Factories{}, err
	}

	// NOTE: include the extension
	extensions := []component.ExtensionFactory{
		configmanagerext.NewFactory(),
	}
	for _, e := range factories.Extensions {
		extensions = append(extensions, e)
	}
	factories.Extensions, err = component.MakeExtensionFactoryMap(extensions...)
	merr.Append(err)

	// NOTE: we might need use our impl of prometheus receiver to calculate custom metrics when scraping (if not already)
	receivers := []component.ReceiverFactory{
		awsecscontainermetricsreceiver.NewFactory(),
	}
	for _, r := range factories.Receivers {
		receivers = append(receivers, r)
	}
	factories.Receivers, err = component.MakeReceiverFactoryMap(receivers...)
	merr.Append(err)

	processors := []component.ProcessorFactory{
		metricstransformprocessor.NewFactory(),
	}
	for _, p := range factories.Processors {
		processors = append(processors, p)
	}
	factories.Processors, err = component.MakeProcessorFactoryMap(processors...)
	merr.Append(err)

	exporters := []component.ExporterFactory{
		awsemfexporter.NewFactory(),
		awsprometheusremotewriteexporter.NewFactory(),
	}
	for _, e := range factories.Exporters {
		exporters = append(exporters, e)
	}
	factories.Exporters, err = component.MakeExporterFactoryMap(exporters...)
	merr.Append(err)

	return factories, merr.ErrorOrNil()
}
