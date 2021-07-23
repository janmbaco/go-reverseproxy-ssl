package certs

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"io/ioutil"
)

// CertificateDefs is the structure that contains the
// public and private key of the certificate and the
// public key of the certificate authority to establish
// secure communication between reverse proxy and virtual host.
type CertificateDefs struct {
	CaPem      string `json:"ca_pem"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// GetCertificate gets the TLS certificate using the public and private key.
func (certificateDefs *CertificateDefs) GetCertificate() tls.Certificate {
	result, err := tls.LoadX509KeyPair(certificateDefs.PublicKey, certificateDefs.PrivateKey)
	errorhandler.TryPanic(err)
	return result
}

// GetTlsConfig gets the config structure to configure a TSL client.
func (certificateDefs *CertificateDefs) GetTlsConfig() *tls.Config {
	tlsConfig := &tls.Config{}
	if len(certificateDefs.CaPem) > 0 {
		tlsConfig = &tls.Config{
			RootCAs: getCertPool(certificateDefs.CaPem),
		}
	}
	//add client certificates
	if len(certificateDefs.PrivateKey) > 0 && len(certificateDefs.PublicKey) > 0 {
		tlsConfig.Certificates = []tls.Certificate{certificateDefs.GetCertificate()}
	}
	return tlsConfig
}

func getCertPool(caPems ...string) *x509.CertPool {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	for _, caPem := range caPems {
		pem, err := ioutil.ReadFile(caPem)
		errorhandler.TryPanic(err)
		rootCAs.AppendCertsFromPEM(pem)
	}
	return rootCAs
}
