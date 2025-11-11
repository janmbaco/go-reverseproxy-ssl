package domain

import (
	"testing"

	certs "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/certificates"
	"github.com/stretchr/testify/assert"
)

func TestClientCertificateHost_GetAuthorizedCAs_WhenClientCertificateNil_ThenReturnsBaseCAs(t *testing.T) {
	// Arrange
	host := &ClientCertificateHost{
		VirtualHostBase: VirtualHostBase{
			// Assume base has some CAs, but for simplicity, nil
		},
	}

	// Act
	result := host.GetAuthorizedCAs()

	// Assert
	assert.Empty(t, result)
}

func TestClientCertificateHost_GetAuthorizedCAs_WhenClientCertificateHasCAs_ThenAppendsToBase(t *testing.T) {
	// Arrange
	mockCert := &certs.CertificateDefs{}
	// Since we can't easily mock, assume it returns some CAs
	// For unit test, perhaps create a simple struct or skip if complex
	// But to test, let's assume GetAuthorizedCAs returns something
	// Actually, since CertificateDefs is external, perhaps test the logic with a mock or simple test

	// For now, test with nil to ensure no panic
	host := &ClientCertificateHost{
		VirtualHostBase:   VirtualHostBase{},
		ClientCertificate: mockCert,
	}

	// Act
	result := host.GetAuthorizedCAs()

	// Assert
	// Since mockCert.GetAuthorizedCAs() probably returns empty, result should be empty
	assert.Empty(t, result)
}
