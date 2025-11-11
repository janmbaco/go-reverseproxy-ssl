package startup

import (
	"testing"

	"github.com/janmbaco/go-infrastructure/v2/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/grpcutil"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigValidator_WhenCalled_ThenReturnsValidator(t *testing.T) {
	// Arrange & Act
	validator := NewConfigValidator()

	// Assert
	assert.NotNil(t, validator)
}

func TestConfigValidator_ValidateRuntime_WhenValidConfig_ThenReturnsTrue(t *testing.T) {
	// Arrange
	validator := NewConfigValidator()
	validConfig := &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "www.example.com",
						Scheme:   "http",
						HostName: "localhost",
						Port:     8080,
					},
				},
			},
		},
		DefaultHost:      "localhost",
		ReverseProxyPort: ":443",
		LogConsoleLevel:  logs.Trace,
		LogFileLevel:     logs.Trace,
		LogsDir:          "./logger",
		ConfigUIPort:     ":8081",
	}

	// Act
	isValid, err := validator.ValidateRuntime(validConfig)

	// Assert
	assert.True(t, isValid)
	assert.NoError(t, err)
}

func TestConfigValidator_ValidateRuntime_WhenInvalidConfig_ThenReturnsFalseWithError(t *testing.T) {
	// Arrange
	validator := NewConfigValidator()
	invalidConfig := &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "",
						Scheme:   "http",
						HostName: "localhost",
						Port:     8080,
					},
				},
			},
		},
		DefaultHost:      "localhost",
		ReverseProxyPort: ":443",
	}

	// Act
	isValid, err := validator.ValidateRuntime(invalidConfig)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'from' field is required")
}

func TestConfigValidator_ValidateRuntime_WhenInvalidType_ThenReturnsFalseWithError(t *testing.T) {
	// Arrange
	validator := NewConfigValidator()
	invalidType := "not a config"

	// Act
	isValid, err := validator.ValidateRuntime(invalidType)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid config type")
}

func TestConfigValidator_ValidateRuntime_WhenValidGrpcWebConfig_ThenReturnsTrue(t *testing.T) {
	// Arrange
	validator := NewConfigValidator()
	config := &domain.Config{
		GrpcWebVirtualHosts: []*domain.GrpcWebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "grpc.example.com",
						Scheme:   "http",
						HostName: "grpc-service",
						Port:     9090,
					},
				},
				GrpcWebProxy: &grpcutil.GrpcWebProxy{
					GrpcProxy: grpcutil.GrpcProxy{
						IsTransparentServer: true,
					},
					AllowAllOrigins: true,
				},
			},
		},
		DefaultHost:      "grpc.example.com",
		ReverseProxyPort: ":443",
		LogConsoleLevel:  logs.Trace,
		LogFileLevel:     logs.Trace,
		LogsDir:          "./logger",
		ConfigUIPort:     ":8081",
	}

	// Act
	isValid, err := validator.ValidateRuntime(config)

	// Assert
	assert.True(t, isValid)
	assert.NoError(t, err)
}
