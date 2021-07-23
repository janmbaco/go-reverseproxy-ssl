package hosts

import (
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
)

// GrpcVirtualHost is used to configure a virtual host using gRPC technology
type GrpcVirtualHost struct {
	ClientCertificateHost
	grpcutil.GrpcProxy
}

func (grpcVirtualHost *GrpcVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	grpcVirtualHost.NewServer(grpcVirtualHost.CreateClientConn(grpcVirtualHost.ClientCertificate, grpcVirtualHost.getHost())).ServeHTTP(rw, req)
}
