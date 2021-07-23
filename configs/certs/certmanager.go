package certs

import (
	"crypto/tls"
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

// AddAutoCertificate registers a virtual host to obtain an automatic Let's encrypt certificate
func (certManager *CertManager) AddAutoCertificate(vhostName string) {
	certManager.autoCertList = append(certManager.autoCertList, vhostName)
	certManager.manager.HostPolicy = autocert.HostWhitelist(certManager.autoCertList...)
}

// AddClientCA registers a list of certificate authorities to use with the reverse proxy
func (certManager *CertManager) AddClientCA(authorizedCA []string) {
	certManager.clientCAs = append(certManager.clientCAs, authorizedCA...)
}

// GetTlsConfig gets the config structure to configure a TSL server
func (certManager *CertManager) GetTlsConfig() *tls.Config {
	ret := &tls.Config{
		GetCertificate: certManager.certificateGetter,
		NextProtos: []string{
			"h2", "http/1.1", // enable HTTP/2
			acme.ALPNProto, // enable tls-alpn ACME challenges
		},
	}
	ret.ClientAuth = tls.VerifyClientCertIfGiven
	ret.ClientCAs = getCertPool(certManager.clientCAs...)
	return ret
}

func (certManager *CertManager) certificateGetter(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if certManager.certificates[hello.ServerName] == nil {
		return certManager.manager.GetCertificate(hello)
	}
	return certManager.certificates[hello.ServerName], nil
}
