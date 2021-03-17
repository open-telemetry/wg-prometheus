package configmanager

import "fmt"

const (
	DefaultHTTPPort       = 8520
	DefaultGRPCPort       = 8521
	DefaultTargetsBaseDir = "/etc/clb"
)

// TODO: use grpc gateway to support both grpc and http
type ServerConfig struct {
	GRPC GRPCServerConfig `mapstructure:"grpc" yaml:"grpc"`
	HTTP HTTPServerConfig `mapstructure:"http" yaml:"http"`
	// Base directory for writing targets files
	TargetsBaseDir string `mapstructure:"targets_base_dir" yaml:"targets_base_dir"`
}

type GRPCServerConfig struct {
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`
}

type HTTPServerConfig struct {
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		GRPC: GRPCServerConfig{
			Endpoint: fmt.Sprintf("0.0.0.0:%d", DefaultGRPCPort),
		},
		HTTP: HTTPServerConfig{
			Endpoint: fmt.Sprintf("0.0.0.0:%d", DefaultHTTPPort),
		},
		TargetsBaseDir: DefaultTargetsBaseDir,
	}
}
