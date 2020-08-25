package hosts

import (
	"encoding/base64"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"net/http"
)

type WebVirtualHost struct {
	*VirtualHost
	ResponseHeaders  map[string]string `json:"response_headers"`
	TlsDefs          *certs.TlsDefs    `json:"tls_defs"`
	NeedPkFromClient bool              `json:"need_pk_from_client"`
}

func (this *WebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if this.NeedPkFromClient && req.TLS.PeerCertificates == nil {
		http.Error(rw, "Not authorized", 401)
		return
	}

	transport := http.DefaultTransport.(*http.Transport)
	if this.TlsDefs != nil {
		transport.TLSClientConfig = this.TlsDefs.GetTlsConfig()
	}

	this.serve(rw, req, func(outReq *http.Request) {
		this.redirectRequest(outReq, req)
		if this.NeedPkFromClient {
			pubKey := base64.URLEncoding.EncodeToString(req.TLS.PeerCertificates[0].RawSubjectPublicKeyInfo)
			outReq.Header.Set("X-Forwarded-ClientKey", pubKey)
		}
	}, transport)
}
