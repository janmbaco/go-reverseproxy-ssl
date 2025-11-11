package domain

import (
	"errors"
	"strconv"
	"strings"

	"github.com/janmbaco/go-infrastructure/v2/logs"
)

// for the reverse proxy, in addition to the various configuration
// Config defines the configuration for the reverse proxy
type Config struct {
	WebVirtualHosts     []*WebVirtualHost     `json:"web_virtual_hosts"`
	GrpcWebVirtualHosts []*GrpcWebVirtualHost `json:"grpc_web_virtual_hosts"`
	DefaultHost         string                `json:"default_host"`
	ReverseProxyPort    string                `json:"reverse_proxy_port"`
	DefaultServerCert   string                `json:"default_server_cert"`
	DefaultServerKey    string                `json:"default_server_key"`
	CertDir             string                `json:"cert_dir"`
	LogConsoleLevel     logs.LogLevel         `json:"log_console_level"`
	LogFileLevel        logs.LogLevel         `json:"log_file_level"`
	LogsDir             string                `json:"logs_dir"`
	ConfigUIPort        string                `json:"config_ui_port"`
	// Deprecated fields for backward compatibility - ignored
	SSHVirtualHosts      interface{} `json:"ssh_virtual_hosts,omitempty"`
	GrpcVirtualHosts     interface{} `json:"grpc_virtual_hosts,omitempty"`
	GrpcJSONVirtualHosts interface{} `json:"grpc_json_virtual_hosts,omitempty"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Check if there are any virtual hosts configured
	if len(c.WebVirtualHosts) == 0 && len(c.GrpcWebVirtualHosts) == 0 {
		return errors.New("at least one virtual host must be configured (web_virtual_hosts or grpc_web_virtual_hosts)")
	}

	// Track domains to ensure uniqueness
	domains := make(map[string]bool)

	// Validate WebVirtualHosts
	for i, host := range c.WebVirtualHosts {
		if err := c.validateVirtualHostBase(&host.VirtualHostBase, i, "web_virtual_hosts"); err != nil {
			return err
		}

		// Validate scheme for web hosts
		if host.Scheme != "http" && host.Scheme != "https" {
			return errors.New("web_virtual_hosts[" + strconv.Itoa(i) + "]: scheme must be 'http' or 'https'")
		}

		// Check domain uniqueness
		if domains[host.From] {
			return errors.New("web_virtual_hosts[" + strconv.Itoa(i) + "]: domain '" + host.From + "' is already used by another virtual host")
		}
		domains[host.From] = true
	}

	// Validate GrpcWebVirtualHosts
	for i, host := range c.GrpcWebVirtualHosts {
		if err := c.validateVirtualHostBase(&host.VirtualHostBase, i, "grpc_web_virtual_hosts"); err != nil {
			return err
		}

		// Validate scheme for grpc-web hosts
		if host.Scheme != "http" && host.Scheme != "https" {
			return errors.New("grpc_web_virtual_hosts[" + strconv.Itoa(i) + "]: scheme must be 'http' or 'https'")
		}

		// Validate required grpc_web_proxy field
		if host.GrpcWebProxy == nil {
			return errors.New("grpc_web_virtual_hosts[" + strconv.Itoa(i) + "]: 'grpc_web_proxy' field is required")
		}

		// Check domain uniqueness
		if domains[host.From] {
			return errors.New("grpc_web_virtual_hosts[" + strconv.Itoa(i) + "]: domain '" + host.From + "' is already used by another virtual host")
		}
		domains[host.From] = true
	}

	// Validate log levels
	if c.LogConsoleLevel < 0 || c.LogConsoleLevel > 5 {
		return errors.New("log_console_level must be between 0 and 5")
	}
	if c.LogFileLevel < 0 || c.LogFileLevel > 5 {
		return errors.New("log_file_level must be between 0 and 5")
	}

	return nil
}

// validateVirtualHostBase validates common virtual host fields
func (c *Config) validateVirtualHostBase(host *VirtualHostBase, index int, arrayName string) error {
	// Validate required fields
	if strings.TrimSpace(host.From) == "" {
		return errors.New(arrayName + "[" + strconv.Itoa(index) + "]: 'from' field is required and cannot be empty")
	}
	if strings.TrimSpace(host.Scheme) == "" {
		return errors.New(arrayName + "[" + strconv.Itoa(index) + "]: 'scheme' field is required and cannot be empty")
	}
	if strings.TrimSpace(host.HostName) == "" {
		return errors.New(arrayName + "[" + strconv.Itoa(index) + "]: 'host_name' field is required and cannot be empty")
	}
	if host.Port == 0 {
		return errors.New(arrayName + "[" + strconv.Itoa(index) + "]: 'port' field is required and must be greater than 0")
	}
	if host.Port > 65535 {
		return errors.New(arrayName + "[" + strconv.Itoa(index) + "]: 'port' field must be between 1 and 65535")
	}

	return nil
}
