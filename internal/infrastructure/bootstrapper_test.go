package infrastructure

import (
	"testing"

	"github.com/janmbaco/go-reverseproxy-ssl/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/internal/infrastructure/grpcutil"
	"github.com/stretchr/testify/assert"
)

func TestNewServerBootstrapper_WhenCalled_ThenReturnsBootstrapper(t *testing.T) {
	// Arrange
	configFile := "test.json"

	// Act
	bootstrapper := NewServerBootstrapper(configFile)

	// Assert
	assert.NotNil(t, bootstrapper)
	assert.Equal(t, configFile, bootstrapper.configFile)
}

func TestServerBootstrapper_createDefaultConfig_WhenCalled_ThenReturnsDefaultConfig(t *testing.T) {
	// Arrange
	bootstrapper := &ServerBootstrapper{}

	// Act
	config := bootstrapper.createDefaultConfig()

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.DefaultHost)
	assert.Equal(t, ":443", config.ReverseProxyPort)
	assert.Equal(t, ":8081", config.ConfigUIPort)
	assert.Len(t, config.WebVirtualHosts, 1)
	assert.Equal(t, "www.example.com", config.WebVirtualHosts[0].From)
}

func TestServerBootstrapper_validateConfig_WhenValidConfig_ThenReturnsTrue(t *testing.T) {
	// Arrange
	bootstrapper := &ServerBootstrapper{}
	validConfig := bootstrapper.createDefaultConfig()

	// Act
	isValid, err := bootstrapper.validateConfig(validConfig)

	// Assert
	assert.True(t, isValid)
	assert.NoError(t, err)
}

func TestServerBootstrapper_validateConfig_WhenInvalidConfig_ThenReturnsFalseWithError(t *testing.T) {
	// Arrange
	bootstrapper := &ServerBootstrapper{}
	invalidConfig := bootstrapper.createDefaultConfig()
	// Make config invalid by clearing required field
	invalidConfig.WebVirtualHosts[0].From = ""

	// Act
	isValid, err := bootstrapper.validateConfig(invalidConfig)

	// Assert
	assert.False(t, isValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'from' field is required")
}

func TestServerBootstrapper_validateConfig_WhenValidGrpcWebConfig_ThenReturnsTrue(t *testing.T) {
	// Arrange
	bootstrapper := &ServerBootstrapper{}
	config := bootstrapper.createDefaultConfig()
	// Replace web host with gRPC-Web host
	config.WebVirtualHosts = nil
	config.GrpcWebVirtualHosts = []*domain.GrpcWebVirtualHost{
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
	}
	config.DefaultHost = "grpc.example.com"

	// Act
	isValid, err := bootstrapper.validateConfig(config)

	// Assert
	assert.True(t, isValid)
	assert.NoError(t, err)
}
