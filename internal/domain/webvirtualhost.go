package domain

import (
	"encoding/base64"
	"net/http"
)

// WebVirtualHost is used to configure a virtual host by web.
type WebVirtualHost struct {
	ClientCertificateHost
	ResponseHeaders  map[string]string `json:"response_headers"`
	NeedPkFromClient bool              `json:"need_pk_from_client"`
}

// WebVirtualHostProvider provides a IVirtualHost
func WebVirtualHostProvider(host *WebVirtualHost, logger Logger) IVirtualHost {
	host.logger = logger
	return host
}

func (webVirtualHost *WebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if webVirtualHost.NeedPkFromClient && req.TLS.PeerCertificates == nil {
		http.Error(rw, "Not authorized", http.StatusUnauthorized)
		return
	}

	transport := http.DefaultTransport.(*http.Transport)
	if webVirtualHost.ClientCertificate != nil {
		tlsConfig, err := webVirtualHost.ClientCertificate.GetTLSConfig()
		if err != nil {
			webVirtualHost.logger.Error("Failed to get TLS config: " + err.Error())
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		transport.TLSClientConfig = tlsConfig
	}

	webVirtualHost.serve(rw, req, func(outReq *http.Request) {
		webVirtualHost.redirectRequest(outReq, req, true)
		if webVirtualHost.NeedPkFromClient {
			pubKey := base64.URLEncoding.EncodeToString(req.TLS.PeerCertificates[0].RawSubjectPublicKeyInfo)
			outReq.Header.Set("X-Forwarded-PrivateKey", pubKey)
		}
	}, transport)
}
