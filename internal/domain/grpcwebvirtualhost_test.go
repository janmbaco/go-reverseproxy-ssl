package domain

import (
	"testing"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/grpcutil"
	"github.com/stretchr/testify/assert"
)

func TestGrpcWebVirtualHost_GrpcWebVirtualHostProvider_WhenCalled_ThenReturnsConfiguredVirtualHost(t *testing.T) {
	// Arrange
	mockLogger := &MockLogger{}
	host := &GrpcWebVirtualHost{
		ClientCertificateHost: ClientCertificateHost{
			VirtualHostBase: VirtualHostBase{
				From: "grpc-web.example.com",
			},
		},
	}
	var mockServer *grpcutil.WrappedGrpcServer

	// Act
	vh := GrpcWebVirtualHostProvider(host, mockServer, mockLogger)

	// Assert
	assert.NotNil(t, vh)
	assert.Equal(t, host, vh)
	assert.Equal(t, mockLogger, host.logger)
}
