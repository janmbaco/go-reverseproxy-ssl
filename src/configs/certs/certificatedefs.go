package certs

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-reverseproxy-ssl/src/configs/certs/certpools"
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
func (certificateDefs *CertificateDefs) GetCertificate() *tls.Certificate {
	result, err := tls.LoadX509KeyPair(certificateDefs.PublicKey, certificateDefs.PrivateKey)
	errorschecker.TryPanic(err)
	return &result
}

// GetTLSConfig gets the config structure to configure a TSL client.
func (certificateDefs *CertificateDefs) GetTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if len(certificateDefs.CaPem) > 0 {
		tlsConfig = &tls.Config{
			RootCAs:    getCertPool(certificateDefs.CaPem...),
			MinVersion: tls.VersionTLS12,
		}
	}
	// add client certificates
	if len(certificateDefs.PrivateKey) > 0 && len(certificateDefs.PublicKey) > 0 {
		tlsConfig.Certificates = []tls.Certificate{*certificateDefs.GetCertificate()}
	}
	return tlsConfig
}

func getCertPool(caPems ...string) *x509.CertPool {
	rootCAs := certpools.GetSysmCertPool()
	for _, caPem := range caPems {
		pem, err := os.ReadFile(caPem)
		errorschecker.TryPanic(err)
		rootCAs.AppendCertsFromPEM(pem)
	}
	return rootCAs
}