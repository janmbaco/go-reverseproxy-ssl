package domain

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/janmbaco/copier"
	"github.com/janmbaco/go-reverseproxy-ssl/internal/infrastructure/grpcutil"
)

// GrpcWebVirtualHost is used to configure a virtual host by grpc web.
type GrpcWebVirtualHost struct {
	ClientCertificateHost
	GrpcWebProxy *grpcutil.GrpcWebProxy `json:"grpc_web_proxy"`
	server       *grpcweb.WrappedGrpcServer
}

// GrpcWebVirtualHostProvider provides a IVirtualHost
func GrpcWebVirtualHostProvider(host *GrpcWebVirtualHost, server *grpcweb.WrappedGrpcServer, logger Logger) IVirtualHost {
	host.server = server
	host.logger = logger
	return host
}

func (g *GrpcWebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var outReq http.Request
	if err := copier.Copy(&outReq, req); err != nil {
		g.logger.Error("Failed to copy request: " + err.Error())
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
	g.redirectRequest(&outReq, req, false)
	g.server.ServeHTTP(rw, &outReq)
}
