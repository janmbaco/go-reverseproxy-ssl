package hosts

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/janmbaco/copier"
	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/src/grpcutil"
)

// GrpcWebVirtualHostTenant defines a tenant name for GrpcWebVirtualHost
const GrpcWebVirtualHostTenant = "GrpcWebVirtualHost"

// GrpcWebVirtualHost is used to configure a virtual host using gRPC-Web technology
type GrpcWebVirtualHost struct {
	ClientCertificateHost
	grpcutil.GrpcWebProxy
	server *grpcweb.WrappedGrpcServer
}

// GrpcWebVirtualHostProvider provides a IVirtualHost
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