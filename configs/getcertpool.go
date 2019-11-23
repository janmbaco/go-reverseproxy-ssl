package configs

import (
	"crypto/x509"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"io/ioutil"
)

func  GetCertPool (caPems ...string) *x509.CertPool{
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	for _, caPem := range caPems{
		pem, err := ioutil.ReadFile(caPem)
		cross.TryPanic(err)
		rootCAs.AppendCertsFromPEM(pem)
	}
	return rootCAs
}
