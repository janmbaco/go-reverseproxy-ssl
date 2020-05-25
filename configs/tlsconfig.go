package configs

import (
	"crypto/tls"
	"golang.org/x/crypto/acme/autocert"
)

func GetTlsConfig(virtualHost []string, caPems []string) *tls.Config {

	autoCert := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache("../certs"),
		HostPolicy: autocert.HostWhitelist(virtualHost...),
	}

	ret := autoCert.TLSConfig()
	ret.ClientAuth = tls.VerifyClientCertIfGiven
	ret.ClientCAs = GetCertPool(caPems...)
	return ret
}
