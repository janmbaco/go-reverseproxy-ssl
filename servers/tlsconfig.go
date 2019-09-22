package servers

import (
	"crypto/rand"
	"crypto/tls"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"golang.org/x/crypto/acme/autocert"
)

func GetTlsConfig(virtualHost []string, caPems []string)  *tls.Config{
	ret := &tls.Config{}
	if Config.DefaultHost == "localhost" {
		//in localhost doesn't works autocert
		cert, err := tls.LoadX509KeyPair("../certs/server.crt", "../certs/server.key")
		cross.TryPanic(err)

		ret = &tls.Config{
			Rand:                  rand.Reader,
			Time:                  nil,
			Certificates:          []tls.Certificate{cert},
			NameToCertificate:     nil,
			GetCertificate:        nil,
			GetClientCertificate:  nil,
			GetConfigForClient:    nil,
			VerifyPeerCertificate: nil,
			RootCAs:               nil,
			NextProtos:            nil,
			ServerName:            "",
			ClientCAs:             nil,
			InsecureSkipVerify:    false,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			PreferServerCipherSuites:    true,
			SessionTicketsDisabled:      false,
			SessionTicketKey:            [32]byte{},
			ClientSessionCache:          nil,
			MinVersion:                  tls.VersionTLS12,
			MaxVersion:                  0,
			CurvePreferences:            []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			DynamicRecordSizingDisabled: false,
			Renegotiation:               0,
			KeyLogWriter:                nil,
		}
	} else {
		autocert := &autocert.Manager{
			Prompt:          autocert.AcceptTOS,
			Cache:           autocert.DirCache("../certs"),
			HostPolicy:      autocert.HostWhitelist(virtualHost...),
		}

		ret = autocert.TLSConfig()
	}
	ret.ClientAuth=tls.VerifyClientCertIfGiven
	ret.ClientCAs = GetCertPool(caPems...)
	return ret
}
