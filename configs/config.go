package configs

import (
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
)

type Config struct {
	VirtualHost      map[string]*VirtualHost `json:"virtual_hosts"`
	DefaultHost      string                  `json:"default_host"`
	ReverseProxyPort string                  `json:"reverse_proxy_port"`
	LogConsoleLevel  cross.LogLevel          `json:"log_console_level"`
	LogFileLevel     cross.LogLevel          `json:"log_file_level"`
	LogsDir          string                  `json:"logs_dir"`
}
