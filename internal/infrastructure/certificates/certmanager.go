package infrastructure

import (
	"crypto/tls"
	"fmt"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

// CertManager is the object responsible for managing the
// certificates globally for the virtual hosts of the reverse proxy
type CertManager struct {
	manager      *autocert.Manager
	autoCertList []string
	clientCAs    []string
	certificates map[string]*tls.Certificate
}

// NewCertManager returns a new object of CertManager type
func NewCertManager(manager *autocert.Manager) *CertManager {
	return &CertManager{manager: manager, autoCertList: make([]string, 0), clientCAs: make([]string, 0), certificates: make(map[string]*tls.Certificate)}
}

// AddCertificate adds a certificate to use on a virtual host
func (certManager *CertManager) AddCertificate(vhostName string, certificate *tls.Certificate) {
	certManager.certificates[vhostName] = certificate
}

// HasCertificateFor indicates if already exists a certificate for de vhostname
func (certManager *CertManager) HasCertificateFor(vhostName string) bool {
	if _, isContained := certManager.certificates[vhostName]; isContained {
		return true
	}
	return false
}

// AddAutoCertificate registers a virtual host to obtain an automatic Let's encrypt certificate
func (certManager *CertManager) AddAutoCertificate(vhostName string) {
	certManager.autoCertList = append(certManager.autoCertList, vhostName)
	certManager.manager.HostPolicy = autocert.HostWhitelist(certManager.autoCertList...)
}

// AddClientCA registers a list of certificate authorities to use with the reverse proxy
func (certManager *CertManager) AddClientCA(authorizedCA []string) {
	certManager.clientCAs = append(certManager.clientCAs, authorizedCA...)
}

// GetTLSConfig gets the config structure to configure a TSL server
func (certManager *CertManager) GetTLSConfig() *tls.Config {
	ret := &tls.Config{
		MinVersion:     tls.VersionTLS12,
		GetCertificate: certManager.certificateGetter,
		NextProtos: []string{
			"h2", "http/1.1", // enable HTTP/2
			acme.ALPNProto, // enable tls-alpn ACME challenges
		},
	}
	ret.ClientAuth = tls.VerifyClientCertIfGiven
	pool, err := getCertPool(certManager.clientCAs...)
	if err != nil {
		return nil
	}
	ret.ClientCAs = pool
	return ret
}

func (certManager *CertManager) certificateGetter(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	// Si tenemos certificado personalizado para este host, usarlo
	if certManager.certificates[hello.ServerName] != nil {
		return certManager.certificates[hello.ServerName], nil
	}

	// Si hay hosts configurados para ACME/Let's Encrypt, usar el manager
	if len(certManager.autoCertList) > 0 {
		return certManager.manager.GetCertificate(hello)
	}

	// No hay certificado para este host y no está configurado ACME
	return nil, fmt.Errorf("no certificate configured for server name: %s", hello.ServerName)
}
