package hosts

import (
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/grpcUtil"
)

type GrpcVirtualHost struct {
	*ClientCertificateHost
	*grpcUtil.GrpcProxy
}

func (this *GrpcVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	this.NewServer(this.CreateClientConn(this.ClientCertificate, this.getHost())).ServeHTTP(rw, req)
}
