package infrastructure

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// CertificateDefs is the structure that contains the
// public and private key of the certificate and the
// public key of the certificate authority to establish
// secure communication between reverse proxy and virtual host.
type CertificateDefs struct {
	CaPem      []string `json:"ca_pems"`
	PublicKey  string   `json:"public_key"`
	PrivateKey string   `json:"private_key"`
}

// GetCertificate gets the TLS certificate using the public and private key.
func (certificateDefs *CertificateDefs) GetCertificate() (*tls.Certificate, error) {
	if certificateDefs == nil {
		return nil, fmt.Errorf("certificate definitions is nil")
	}
	if certificateDefs.PublicKey == "" || certificateDefs.PrivateKey == "" {
		return nil, fmt.Errorf("certificate public key or private key is empty")
	}
	result, err := tls.LoadX509KeyPair(certificateDefs.PublicKey, certificateDefs.PrivateKey)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTLSConfig gets the config structure to configure a TSL client.
func (certificateDefs *CertificateDefs) GetTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if len(certificateDefs.CaPem) > 0 {
		rootCAs, err := getCertPool(certificateDefs.CaPem...)
		if err != nil {
			return nil, err
		}
		tlsConfig = &tls.Config{
			RootCAs:    rootCAs,
			MinVersion: tls.VersionTLS12,
		}
	}
	// add client certificates
	if len(certificateDefs.PrivateKey) > 0 && len(certificateDefs.PublicKey) > 0 {
		cert, err := certificateDefs.GetCertificate()
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}
	return tlsConfig, nil
}

// GetAuthorizedCAs gets the certificate authorities public keys.
func (certificateDefs *CertificateDefs) GetAuthorizedCAs() []string {
	return certificateDefs.CaPem
}

func getCertPool(caPems ...string) (*x509.CertPool, error) {
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	for _, caPem := range caPems {
		pem, err := os.ReadFile(caPem)
		if err != nil {
			return nil, err
		}
		rootCAs.AppendCertsFromPEM(pem)
	}
	return rootCAs, nil
}
