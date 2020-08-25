package certs

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
)

type TlsDefs struct {
	CaPem     string `json:"ca_pem"`
	ClientCrt string `json:"client_crt"`
	ClientKey string `json:"client_key"`
}

func (this *TlsDefs) GetTlsConfig() *tls.Config {
	tlsConfig := &tls.Config{}
	if len(this.CaPem) > 0 {
		tlsConfig = &tls.Config{
			RootCAs: getCertPool(this.CaPem),
		}
	}
	//add client certificates
	if len(this.ClientKey) > 0 && len(this.ClientCrt) > 0 {
		clientCert, err := tls.LoadX509KeyPair(this.ClientCrt, this.ClientKey)
		errorhandler.TryPanic(err)
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}
	return tlsConfig
}

func GetAutoCertConfig(virtualHost []string, caPems []string) *tls.Config {

	autoCert := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("../certs"),
		HostPolicy: autocert.HostWhitelist(virtualHost...),
	}

	ret := autoCert.TLSConfig()
	ret.ClientAuth = tls.VerifyClientCertIfGiven
	ret.ClientCAs = getCertPool(caPems...)
	return ret
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
