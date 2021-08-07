package configs

import (
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
)

// A Config structure is used to configure multiple virtual hosts
// for the reverse proxy, in addition to the various configuration
// options for the reverse proxy.
type Config struct {
	WebVirtualHosts      []*hosts.WebVirtualHost      `json:"web_virtual_hosts"`
	SshVirtualHosts      []*hosts.SSHVirtualHost      `json:"ssh_virtual_hosts"`
	GrpcVirtualHosts     []*hosts.GrpcVirtualHost     `json:"grpc_virtual_hosts"`
	GrpcJSONVirtualHosts []*hosts.GrpcJSONVirtualHost `json:"grpc_json_virtual_hosts"`
	GrpcWebVirtualHosts  []*hosts.GrpcWebVirtualHost  `json:"grpc_web_virtual_hosts"`
	DefaultHost          string                       `json:"default_host"`
	ReverseProxyPort     string                       `json:"reverse_proxy_port"`
	LogConsoleLevel      logs.LogLevel                `json:"log_console_level"`
	LogFileLevel         logs.LogLevel                `json:"log_file_level"`
	LogsDir              string                       `json:"logs_dir"`
}
