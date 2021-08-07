// +build windows

package certpools

import (
	"crypto/x509"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	CryptENotFound = 0x80092004
)

func GetSysmCertPool() *x509.CertPool {
	storeHandle, err := syscall.CertOpenSystemStore(0, syscall.StringToUTF16Ptr("Root"))
	if err != nil {
		fmt.Println(syscall.GetLastError())
	}

	var certs []*x509.Certificate
	var cert *syscall.CertContext
	for {
		cert, err = syscall.CertEnumCertificatesInStore(storeHandle, cert)
		if err != nil {
			var errno syscall.Errno
			if ok := errors.Is(err, errno); ok {
				if errno == CryptENotFound {
					break
				}
			}
			fmt.Println(syscall.GetLastError())
		}
		if cert == nil {
			break
		}
		// Copy the buf, since ParseCertificate does not create its own copy.
		buf := (*[1 << 20]byte)(unsafe.Pointer(cert.EncodedCert))[:]
		buf2 := make([]byte, cert.Length)
		copy(buf2, buf)
		if c, err := x509.ParseCertificate(buf2); err == nil {
			certs = append(certs, c)
		}
	}
	certPool := x509.NewCertPool()

	for _, cert := range certs {
		certPool.AddCert(cert)
	}

	return certPool
}
