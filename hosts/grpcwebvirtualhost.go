package hosts

import (
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/grpcUtil"
)

type GrpcWebVirtualHost struct {
	ClientCertificateHost
	grpcUtil.GrpcWebProxy
}

func (this *GrpcWebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	this.WrappedGrpcServer(this.CreateClientConn(this.ClientCertificate, this.getHost())).ServeHTTP(rw, req)
}
