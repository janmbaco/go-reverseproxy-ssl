package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebVirtualHost_WebVirtualHostProvider_WhenCalled_ThenReturnsConfiguredVirtualHost(t *testing.T) {
	// Arrange
	mockLogger := &MockLogger{}
	host := &WebVirtualHost{
		ClientCertificateHost: ClientCertificateHost{
			VirtualHostBase: VirtualHostBase{
				From: "web.example.com",
			},
		},
		ResponseHeaders:  map[string]string{"X-Custom": "value"},
		NeedPkFromClient: true,
	}

	// Act
	vh := WebVirtualHostProvider(host, mockLogger)

	// Assert
	assert.NotNil(t, vh)
	assert.Equal(t, host, vh)
	assert.Equal(t, mockLogger, host.logger)
}
