package hosts

import "github.com/janmbaco/go-reverseproxy-ssl/configs/certs"

type ClientCertificateHost struct {
	*VirtualHost
	ClientCertificate *certs.CertificateDefs `json:"client_certificate"`
}

func (this *ClientCertificateHost) GetAuthorizedCAs() []string {
	CAs := this.VirtualHost.GetAuthorizedCAs()
	if this.ClientCertificate != nil && len(this.ClientCertificate.CaPem) > 0 {
		CAs = append(CAs, this.ClientCertificate.CaPem)
	}
	return CAs
}
