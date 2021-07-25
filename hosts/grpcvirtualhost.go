package hosts

import (
	"github.com/janmbaco/copier"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
)

// GrpcVirtualHost is used to configure a virtual host using gRPC technology
type GrpcVirtualHost struct {
	ClientCertificateHost
	grpcutil.GrpcProxy
}

func (grpcVirtualHost *GrpcVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var outReq http.Request
	errorhandler.TryPanic(copier.Copy(&outReq, req))
	grpcVirtualHost.redirectRequest(&outReq, req, false)
	grpcVirtualHost.NewServer(grpcVirtualHost.CreateClientConn(grpcVirtualHost.ClientCertificate, grpcVirtualHost.getHost())).ServeHTTP(rw, &outReq)
}
