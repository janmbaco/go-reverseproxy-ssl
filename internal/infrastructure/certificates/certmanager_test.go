package infrastructure

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/acme/autocert"
)

func TestNewCertManager_WhenCalled_ThenReturnsManager(t *testing.T) {
	// Arrange
	manager := &autocert.Manager{}

	// Act
	certMgr := NewCertManager(manager)

	// Assert
	assert.NotNil(t, certMgr)
	assert.Equal(t, manager, certMgr.manager)
	assert.Empty(t, certMgr.autoCertList)
	assert.Empty(t, certMgr.clientCAs)
	assert.Empty(t, certMgr.certificates)
}

func TestCertManager_AddCertificate_WhenCalled_ThenAddsCertificate(t *testing.T) {
	// Arrange
	certMgr := NewCertManager(&autocert.Manager{})
	cert := &tls.Certificate{}

	// Act
	certMgr.AddCertificate("example.com", cert)

	// Assert
	assert.True(t, certMgr.HasCertificateFor("example.com"))
	assert.False(t, certMgr.HasCertificateFor("other.com"))
}

func TestCertManager_HasCertificateFor_WhenCertificateExists_ThenReturnsTrue(t *testing.T) {
	// Arrange
	certMgr := NewCertManager(&autocert.Manager{})
	cert := &tls.Certificate{}
	certMgr.AddCertificate("example.com", cert)

	// Act
	result := certMgr.HasCertificateFor("example.com")

	// Assert
	assert.True(t, result)
}

func TestCertManager_AddAutoCertificate_WhenCalled_ThenAddsToList(t *testing.T) {
	// Arrange
	certMgr := NewCertManager(&autocert.Manager{})

	// Act
	certMgr.AddAutoCertificate("example.com")

	// Assert
	assert.Contains(t, certMgr.autoCertList, "example.com")
}

func TestCertManager_AddClientCA_WhenCalled_ThenAddsToList(t *testing.T) {
	// Arrange
	certMgr := NewCertManager(&autocert.Manager{})

	// Act
	certMgr.AddClientCA([]string{"ca1.pem", "ca2.pem"})

	// Assert
	assert.Contains(t, certMgr.clientCAs, "ca1.pem")
	assert.Contains(t, certMgr.clientCAs, "ca2.pem")
}

func TestCertManager_GetTLSConfig_WhenCalled_ThenReturnsConfig(t *testing.T) {
	// Arrange
	certMgr := NewCertManager(&autocert.Manager{})

	// Act
	config := certMgr.GetTLSConfig()

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, uint16(771), config.MinVersion) // tls.VersionTLS12
	assert.Equal(t, tls.VerifyClientCertIfGiven, config.ClientAuth)
	assert.Contains(t, config.NextProtos, "h2")
	assert.Contains(t, config.NextProtos, "http/1.1")
	assert.Contains(t, config.NextProtos, "acme-tls/1")
}
