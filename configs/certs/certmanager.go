package certs

import (
	"crypto/tls"
	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
)

type CertManager struct {
	manager      *autocert.Manager
	autoCertList []string
	clientCAs    []string
	certificates map[string]*tls.Certificate
}

func NewCertManager(manager *autocert.Manager) *CertManager {
	return &CertManager{manager: manager, autoCertList: make([]string, 0), clientCAs: make([]string, 0), certificates: make(map[string]*tls.Certificate)}
}

func (this *CertManager) AddCertificate(vhostName string, certificate *tls.Certificate) {
	this.certificates[vhostName] = certificate
}

func (this *CertManager) AddAutoCertificate(vhostName string) {
	this.autoCertList = append(this.autoCertList, vhostName)
	this.manager.HostPolicy = autocert.HostWhitelist(this.autoCertList...)
}

func (this *CertManager) AddClientCA(authorizedCA []string) {
	this.clientCAs = append(this.clientCAs, authorizedCA...)
}

func (this *CertManager) GetTlsConfig() *tls.Config {
	ret := &tls.Config{
		GetCertificate: this.GetCertificate,
		NextProtos: []string{
			"h2", "http/1.1", // enable HTTP/2
			acme.ALPNProto, // enable tls-alpn ACME challenges
		},
	}
	ret.ClientAuth = tls.VerifyClientCertIfGiven
	ret.ClientCAs = getCertPool(this.clientCAs...)
	return ret
}

func (this *CertManager) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
	if this.certificates[hello.ServerName] == nil {
		return this.manager.GetCertificate(hello)
	}
	return this.certificates[hello.ServerName], nil
}
