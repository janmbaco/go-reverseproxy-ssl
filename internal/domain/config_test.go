package domain

import (
	"encoding/json"
	"testing"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/grpcutil"
	"github.com/stretchr/testify/assert"
)

func TestConfig_BackwardCompatibility_WhenDeprecatedFieldsPresent_ThenUnmarshalsSuccessfully(t *testing.T) {
	// Arrange
	configJSON := `{
		"web_virtual_hosts": [],
		"grpc_web_virtual_hosts": [],
		"ssh_virtual_hosts": [{"from": "old-ssh.example.com"}],
		"grpc_json_virtual_hosts": [{"from": "old-grpc-json.example.com"}],
		"default_host": "localhost",
		"reverse_proxy_port": "8080",
		"default_server_cert": "",
		"default_server_key": "",
		"cert_dir": "",
		"log_console_level": 1,
		"log_file_level": 1,
		"logs_dir": "",
		"config_ui_port": "8081"
	}`

	// Act
	var config Config
	err := json.Unmarshal([]byte(configJSON), &config)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, config.WebVirtualHosts, 0)
	assert.Len(t, config.GrpcWebVirtualHosts, 0)
	assert.Equal(t, "localhost", config.DefaultHost)
}

func TestConfig_Validate_WhenValidGrpcWebConfig_ThenReturnsNil(t *testing.T) {
	// Arrange
	config := &Config{
		GrpcWebVirtualHosts: []*GrpcWebVirtualHost{
			{
				ClientCertificateHost: ClientCertificateHost{
					VirtualHostBase: VirtualHostBase{
						From:     "grpc.example.com",
						Scheme:   "http",
						HostName: "grpc-service",
						Port:     9090,
					},
				},
				GrpcWebProxy: &grpcutil.GrpcWebProxy{
					GrpcProxy: grpcutil.GrpcProxy{
						GrpcServices: map[string][]string{
							"hello.HelloService": {"SayHello"},
						},
						IsTransparentServer: true,
						Authority:           "grpc-service:9090",
					},
					AllowAllOrigins: true,
					AllowedOrigins:  []string{},
					UseWebSockets:   false,
					AllowedHeaders:  []string{},
				},
			},
		},
		DefaultHost:      "grpc.example.com",
		ReverseProxyPort: ":443",
		LogConsoleLevel:  4,
		LogFileLevel:     4,
	}

	// Act
	err := config.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestConfig_Validate_WhenNoVirtualHosts_ThenReturnsError(t *testing.T) {
	// Arrange
	config := &Config{
		WebVirtualHosts:     []*WebVirtualHost{},
		GrpcWebVirtualHosts: []*GrpcWebVirtualHost{},
	}

	// Act
	err := config.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one virtual host must be configured")
}

func TestConfig_Validate_WhenMissingRequiredFields_ThenReturnsError(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected string
	}{
		{
			name: "empty from field",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "",
								Scheme:   "http",
								HostName: "localhost",
								Port:     8080,
							},
						},
					},
				},
			},
			expected: "'from' field is required and cannot be empty",
		},
		{
			name: "empty scheme field",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "",
								HostName: "localhost",
								Port:     8080,
							},
						},
					},
				},
			},
			expected: "'scheme' field is required and cannot be empty",
		},
		{
			name: "empty host_name field",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "http",
								HostName: "",
								Port:     8080,
							},
						},
					},
				},
			},
			expected: "'host_name' field is required and cannot be empty",
		},
		{
			name: "zero port",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "http",
								HostName: "localhost",
								Port:     0,
							},
						},
					},
				},
			},
			expected: "'port' field is required and must be greater than 0",
		},
		{
			name: "invalid port range",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "http",
								HostName: "localhost",
								Port:     70000,
							},
						},
					},
				},
			},
			expected: "'port' field must be between 1 and 65535",
		},
		{
			name: "invalid scheme for web host",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "tcp",
								HostName: "localhost",
								Port:     8080,
							},
						},
					},
				},
			},
			expected: "scheme must be 'http' or 'https'",
		},
		{
			name: "duplicate domains",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "http",
								HostName: "localhost",
								Port:     8080,
							},
						},
					},
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "http",
								HostName: "localhost",
								Port:     8081,
							},
						},
					},
				},
			},
			expected: "domain 'example.com' is already used by another virtual host",
		},
		{
			name: "invalid log levels",
			config: &Config{
				WebVirtualHosts: []*WebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "example.com",
								Scheme:   "http",
								HostName: "localhost",
								Port:     8080,
							},
						},
					},
				},
				LogConsoleLevel: -1,
				LogFileLevel:    6,
			},
			expected: "log_console_level must be between 0 and 5",
		},
		{
			name: "missing grpc_web_proxy in grpc_web_virtual_host",
			config: &Config{
				GrpcWebVirtualHosts: []*GrpcWebVirtualHost{
					{
						ClientCertificateHost: ClientCertificateHost{
							VirtualHostBase: VirtualHostBase{
								From:     "grpc.example.com",
								Scheme:   "http",
								HostName: "grpc-service",
								Port:     9090,
							},
						},
						GrpcWebProxy: nil, // Missing
					},
				},
			},
			expected: "'grpc_web_proxy' field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := tt.config.Validate()

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expected)
		})
	}
}
