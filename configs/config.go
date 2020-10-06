package configs

import (
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
)

type Config struct {
	WebVirtualHosts      map[string]*hosts.WebVirtualHost      `json:"web_virtual_hosts"`
	SshVirtualHosts      map[string]*hosts.SshVirtualHost      `json:"ssh_virtual_hosts"`
	GrpcVirtualHosts     map[string]*hosts.GrpcVirtualHost     `json:"grpc_virtual_hosts"`
	GrpcJsonVirtualHosts map[string]*hosts.GrpcJsonVirtualHost `json:"grpc_json_virtual_hosts"`
	GrpcWebVirtualHosts  map[string]*hosts.GrpcWebVirtualHost  `json:"grpc_web_virtual_hosts"`
	DefaultHost          string                                `json:"default_host"`
	ReverseProxyPort     string                                `json:"reverse_proxy_port"`
	LogConsoleLevel      logs.LogLevel                         `json:"log_console_level"`
	LogFileLevel         logs.LogLevel                         `json:"log_file_level"`
	LogsDir              string                                `json:"logs_dir"`
}
