package hosts

import "github.com/janmbaco/go-reverseproxy-ssl/src/configs/certs"

// ClientCertificateHost is used to configure a simple virtual host using TLS client communication.
type ClientCertificateHost struct {
	VirtualHost
	ClientCertificate *certs.CertificateDefs `json:"client_certificate"`
}

// GetAuthorizedCAs gets the certificate authorities public keys to use with TLS client.
func (clientCertificateHost *ClientCertificateHost) GetAuthorizedCAs() []string {
	CAs := clientCertificateHost.VirtualHost.GetAuthorizedCAs()
	if clientCertificateHost.ClientCertificate != nil && len(clientCertificateHost.ClientCertificate.CaPem) > 0 {
		CAs = append(CAs, clientCertificateHost.ClientCertificate.CaPem...)
	}
	return CAs
}