// +build linux netbsd openbsd solaris freebsd darwin

package certpools

import (
	"crypto/x509"
)

func GetSysmCertPool() *x509.CertPool {
	certPool, _ := x509.SystemCertPool()
	return certPool
}
