package certs

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"io/ioutil"
)

type CertificateDefs struct {
	CaPem      string `json:"ca_pem"`
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

func (this *CertificateDefs) GetCertificate() tls.Certificate {
	result, err := tls.LoadX509KeyPair(this.PublicKey, this.PrivateKey)
	errorhandler.TryPanic(err)
	return result
}

func (this *CertificateDefs) GetTlsConfig() *tls.Config {
	tlsConfig := &tls.Config{}
	if len(this.CaPem) > 0 {
		tlsConfig = &tls.Config{
			RootCAs: getCertPool(this.CaPem),
		}
	}
	//add client certificates
	if len(this.PrivateKey) > 0 && len(this.PublicKey) > 0 {
		tlsConfig.Certificates = []tls.Certificate{this.GetCertificate()}
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
