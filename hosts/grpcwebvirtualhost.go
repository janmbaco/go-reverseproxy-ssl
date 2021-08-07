package hosts

import (
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/janmbaco/copier"
	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
	"net/http"
)

const GrpcWebVirtualHostTenant = "GrpcWebVirtualHost"

// GrpcWebVirtualHost is used to configure a virtual host using gRPC-Web technology
type GrpcWebVirtualHost struct {
	ClientCertificateHost
	grpcutil.GrpcWebProxy
	server *grpcweb.WrappedGrpcServer
}

func GrpcWebVirtualHostProvider(host *GrpcWebVirtualHost, server *grpcweb.WrappedGrpcServer, logger logs.Logger) IVirtualHost {
	host.server = server
	host.logger = logger
	return host
}

func (grpcWebVirtualHost *GrpcWebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var outReq http.Request
	errorschecker.TryPanic(copier.Copy(&outReq, req))
	grpcWebVirtualHost.redirectRequest(&outReq, req, false)
	grpcWebVirtualHost.server.ServeHTTP(rw, &outReq)
}
