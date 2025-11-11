package domain

import (
	certs "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/certificates"
)

// ClientCertificateHost is used to configure a simple virtual host using TLS client communication.
type ClientCertificateHost struct {
	VirtualHostBase
	ClientCertificate *certs.CertificateDefs `json:"client_certificate"`
}

// GetAuthorizedCAs gets the certificate authorities public keys to use with TLS client.
func (clientCertificateHost *ClientCertificateHost) GetAuthorizedCAs() []string {
	CAs := clientCertificateHost.VirtualHostBase.GetAuthorizedCAs()
	if clientCertificateHost.ClientCertificate != nil && len(clientCertificateHost.ClientCertificate.GetAuthorizedCAs()) > 0 {
		CAs = append(CAs, clientCertificateHost.ClientCertificate.GetAuthorizedCAs()...)
	}
	return CAs
}
