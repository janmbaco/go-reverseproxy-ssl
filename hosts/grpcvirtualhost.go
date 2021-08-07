package hosts

import (
	"github.com/janmbaco/copier"
	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
	"google.golang.org/grpc"
	"net/http"
)

const GrpcVirtualHostTenant = "GrpcVirtualHost"

// GrpcVirtualHost is used to configure a virtual host using gRPC technology
type GrpcVirtualHost struct {
	ClientCertificateHost
	grpcutil.GrpcProxy
	server *grpc.Server
}

func GrpcVirtualHostProvider(host *GrpcVirtualHost, server *grpc.Server, logger logs.Logger) IVirtualHost {
	host.server = server
	host.logger = logger
	return host
}

func (grpcVirtualHost *GrpcVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var outReq http.Request
	errorschecker.TryPanic(copier.Copy(&outReq, req))
	grpcVirtualHost.redirectRequest(&outReq, req, false)
	grpcVirtualHost.server.ServeHTTP(rw, &outReq)
}
