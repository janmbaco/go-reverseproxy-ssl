package hosts

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/src/grpcutil"
)

// GrpcJSONVirtualHostTenant defines a tenant name for GrpcJSONVirtualHost
const GrpcJSONVirtualHostTenant = "GrpcJsonVirtualHost"

// GrpcJSONVirtualHost is used to configure a virtual host with a web client (json) and a gRPC server.
type GrpcJSONVirtualHost struct {
	ClientCertificateHost
	transport grpcutil.TransportJSON
}

// GrpcJSONVirtualHostProvider provides a IVirtualHost
func GrpcJSONVirtualHostProvider(host *GrpcJSONVirtualHost, trnsport grpcutil.TransportJSON, logger logs.Logger) IVirtualHost {
	host.transport = trnsport
	host.logger = logger
	return host
}

func (grpcJsonVirtualHost *GrpcJSONVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	_, err := url.Parse(grpcJsonVirtualHost.Scheme + "://" + grpcJsonVirtualHost.HostName + ":" + strconv.Itoa(int(grpcJsonVirtualHost.Port)))
	errorschecker.TryPanic(err)
	grpcJsonVirtualHost.serve(rw, req, func(outReq *http.Request) {
		grpcJsonVirtualHost.redirectRequest(outReq, req, true)
	}, grpcJsonVirtualHost.transport)
}