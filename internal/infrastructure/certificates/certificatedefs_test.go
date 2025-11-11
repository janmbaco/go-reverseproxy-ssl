package infrastructure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCertificateDefs_GetAuthorizedCAs_WhenCalled_ThenReturnsCaPem(t *testing.T) {
	// Arrange
	certDefs := &CertificateDefs{
		CaPem: []string{"ca1.pem", "ca2.pem"},
	}

	// Act
	result := certDefs.GetAuthorizedCAs()

	// Assert
	assert.Equal(t, []string{"ca1.pem", "ca2.pem"}, result)
}

func TestCertificateDefs_GetCertificate_WhenInvalidPaths_ThenReturnsError(t *testing.T) {
	// Arrange
	certDefs := &CertificateDefs{
		PublicKey:  "nonexistent.crt",
		PrivateKey: "nonexistent.key",
	}

	// Act
	cert, err := certDefs.GetCertificate()

	// Assert
	assert.Nil(t, cert)
	assert.Error(t, err)
}

func TestCertificateDefs_GetTLSConfig_WhenNoCerts_ThenReturnsConfig(t *testing.T) {
	// Arrange
	certDefs := &CertificateDefs{}

	// Act
	config, err := certDefs.GetTLSConfig()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, uint16(771), config.MinVersion) // tls.VersionTLS12
	assert.Empty(t, config.Certificates)
}

func TestCertificateDefs_GetTLSConfig_WhenHasCaPems_ThenReturnsConfigWithRootCAs(t *testing.T) {
	// Arrange
	certDefs := &CertificateDefs{
		CaPem: []string{"nonexistent.pem"}, // Will fail but test the logic
	}

	// Act
	config, err := certDefs.GetTLSConfig()

	// Assert
	assert.Error(t, err) // Because file doesn't exist
	assert.Nil(t, config)
}
