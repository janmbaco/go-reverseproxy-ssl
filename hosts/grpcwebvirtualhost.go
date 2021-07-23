package hosts

import (
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
)

// GrpcWebVirtualHost is used to configure a virtual host using gRPC-Web technology
type GrpcWebVirtualHost struct {
	ClientCertificateHost
	grpcutil.GrpcWebProxy
}

func (grpcWebVirtualHost *GrpcWebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	grpcWebVirtualHost.WrappedGrpcServer(grpcWebVirtualHost.CreateClientConn(grpcWebVirtualHost.ClientCertificate, grpcWebVirtualHost.getHost())).ServeHTTP(rw, req)
}
