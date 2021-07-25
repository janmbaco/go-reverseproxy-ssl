package configs

import (
	"github.com/janmbaco/go-infrastructure/config"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
)

// A Config structure is used to configure multiple virtual hosts
// for the reverse proxy, in addition to the various configuration
// options for the reverse proxy.
type Config struct {
	config.ConfigHandler `json:"-"`
	WebVirtualHosts      []*hosts.WebVirtualHost      `json:"web_virtual_hosts"`
	SshVirtualHosts      []*hosts.SshVirtualHost      `json:"ssh_virtual_hosts"`
	GrpcVirtualHosts     []*hosts.GrpcVirtualHost     `json:"grpc_virtual_hosts"`
	GrpcJsonVirtualHosts []*hosts.GrpcJsonVirtualHost `json:"grpc_json_virtual_hosts"`
	GrpcWebVirtualHosts  []*hosts.GrpcWebVirtualHost  `json:"grpc_web_virtual_hosts"`
	DefaultHost          string                       `json:"default_host"`
	ReverseProxyPort     string                       `json:"reverse_proxy_port"`
	LogConsoleLevel      logs.LogLevel                `json:"log_console_level"`
	LogFileLevel         logs.LogLevel                `json:"log_file_level"`
	LogsDir              string                       `json:"logs_dir"`
}

func NewConfig(configHandler config.ConfigHandler, defaultHost string, reverseProxyPort string, logConsoleLevel logs.LogLevel, logFileLevel logs.LogLevel, logsDir string) *Config {
	result := &Config{ConfigHandler: configHandler, DefaultHost: defaultHost, ReverseProxyPort: reverseProxyPort, LogConsoleLevel: logConsoleLevel, LogFileLevel: logFileLevel, LogsDir: logsDir}
	result.Load(result)
	return result
}

func NewConfigHandler(file string) config.ConfigHandler {
	return config.NewFileConfigHandler(file)
}
