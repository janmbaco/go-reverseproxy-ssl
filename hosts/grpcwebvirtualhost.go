package hosts

import (
	"github.com/janmbaco/copier"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
)

// GrpcWebVirtualHost is used to configure a virtual host using gRPC-Web technology
type GrpcWebVirtualHost struct {
	ClientCertificateHost
	grpcutil.GrpcWebProxy
}

func (grpcWebVirtualHost *GrpcWebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var outReq http.Request
	errorhandler.TryPanic(copier.Copy(&outReq, req))
	grpcWebVirtualHost.redirectRequest(&outReq, req, false)
	grpcWebVirtualHost.WrappedGrpcServer(grpcWebVirtualHost.CreateClientConn(grpcWebVirtualHost.ClientCertificate, grpcWebVirtualHost.getHost())).ServeHTTP(rw, &outReq)
}
