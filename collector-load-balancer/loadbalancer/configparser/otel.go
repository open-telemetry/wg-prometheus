package configparser

import (
	"path/filepath"
	"strings"

	"github.com/dyweb/gommon/errors"
	"gopkg.in/yaml.v2"
)

type OtelConfig struct {
	Extensions interface{}                       `yaml:"extensions"`
	Receivers  map[string]map[string]interface{} `yaml:"receivers"`
	Processors interface{}                       `yaml:"processors"`
	Exporters  interface{}                       `yaml:"exporters"`
	Service    interface{}                       `yaml:"service"`
}

const (
	k8sSDConfigKey        = "kubernetes_sd_configs"
	fileSDRefreshInterval = "1m"
)

type PrometheusConfig struct {
	ScrapeConfigs []map[string]interface{} `yaml:"scrape_configs"`
}

type PrometheusFileSDConfig struct {
	Files           []string `yaml:"files"`
	RefreshInterval string   `yaml:"refresh_interval"`
}

// ExtractPrometheusK8sSD returns a full prometheus config but remove all the service discovery except k8s.
// static config is also removed. It is for running prometheus k8s sd on a central server.
func ExtractPrometheusK8sSD(otelCfg []byte) ([]byte, error) {
	// Look for prometheus receiver config
	var cfg OtelConfig
	if err := yaml.Unmarshal(otelCfg, &cfg); err != nil {
		return nil, errors.Wrap(err, "decode otel yaml failed")
	}
	// TODO: there can be multiple receivers like prometheus/1
	promReceiver, ok := cfg.Receivers["prometheus"]
	if !ok {
		return nil, errors.New("prometheus config not found")
	}
	promCfg, ok := promReceiver["config"]
	if !ok {
		return nil, errors.New("config within prometheus not found")
	}

	// Encoded and decode to 'cast' interface{} into PrometheusConfig ...
	promStr, err := yaml.Marshal(promCfg)
	if err != nil {
		return nil, errors.Wrap(err, "encode prometheus section in otel config failed")
	}
	var k8sOnlyConfig PrometheusConfig
	if err := yaml.Unmarshal(promStr, &k8sOnlyConfig); err != nil {
		return nil, errors.Wrap(err, "decode scrape config failed")
	}
	// Remove all sd configs except k8s
	for _, scrapeConfig := range k8sOnlyConfig.ScrapeConfigs {
		var keysToRemove []string
		for k, _ := range scrapeConfig {
			switch {
			case k == k8sSDConfigKey:
				continue
			case strings.HasSuffix(k, "_sd_configs"):
				fallthrough
			case k == "static_configs":
				keysToRemove = append(keysToRemove, k)
			}
		}
		for _, k := range keysToRemove {
			delete(scrapeConfig, k)
		}
		// TODO: maybe just remove the job if all the sd are removed ...
	}
	promStr, err = yaml.Marshal(k8sOnlyConfig)
	if err != nil {
		return nil, errors.Wrap(err, "encode prometheus section in otel config failed")
	}
	return promStr, nil
}

// ReplacePrometheusK8sSD replace all k8s sd config to file sd inside otel config
func ReplacePrometheusK8sSD(otelCfg []byte, targetsBaseDir string) ([]byte, error) {
	var cfg OtelConfig
	if err := yaml.Unmarshal(otelCfg, &cfg); err != nil {
		return nil, errors.Wrap(err, "decode otel yaml failed")
	}
	// TODO: there can be multiple receivers like prometheus/1
	promReceiver, ok := cfg.Receivers["prometheus"]
	if !ok {
		return nil, errors.New("prometheus config not found")
	}
	promCfg, ok := promReceiver["config"]
	if !ok {
		return nil, errors.New("config within prometheus not found")
	}

	// 'cast'
	promStr, err := yaml.Marshal(promCfg)
	if err != nil {
		return nil, errors.Wrap(err, "encode prometheus section in otel config failed")
	}
	var replacedPromCfg PrometheusConfig
	if err := yaml.Unmarshal(promStr, &replacedPromCfg); err != nil {
		return nil, errors.Wrap(err, "decode scrape config failed")
	}
	// Replace k8s with file sd
	for _, scrapeConfig := range replacedPromCfg.ScrapeConfigs {
		var jobName string
		hasK8s := false
		for k, v := range scrapeConfig {
			switch k {
			case "job_name":
				jobName = v.(string)
			case k8sSDConfigKey:
				hasK8s = true
			}
		}
		if hasK8s {
			delete(scrapeConfig, "kubernetes_sd_configs")
			// FIXME: we are ignoring existing file sd configs, we should do another 'cast' and append
			scrapeConfig["file_sd_configs"] = []PrometheusFileSDConfig{
				{
					Files: []string{
						filepath.Join(targetsBaseDir, jobName+".json"),
					},
					RefreshInterval: fileSDRefreshInterval,
				},
			}
		}
	}

	// Use the replace config in otel config
	promReceiver["config"] = replacedPromCfg
	otelCfg, err = yaml.Marshal(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "encode otel config failed")
	}
	return otelCfg, err
}
