package configs

import (
	"crypto/x509"
	"io/ioutil"

	"github.com/janmbaco/go-infrastructure/errorhandler"
)

func GetCertPool(caPems ...string) *x509.CertPool {
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
