package hosts

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/janmbaco/go-infrastructure/errorhandler"
	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
)

// GrpcJsonVirtualHost is used to configure a virtual host with a web client (json) and a gRPC server.
type GrpcJsonVirtualHost struct {
	ClientCertificateHost
}

func (grpcJsonVirtualHost *GrpcJsonVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	_, err := url.Parse(grpcJsonVirtualHost.Scheme + "://" + grpcJsonVirtualHost.HostName + ":" + strconv.Itoa(int(grpcJsonVirtualHost.Port)))
	errorhandler.TryPanic(err)
	grpcJsonVirtualHost.serve(rw, req, func(outReq *http.Request) {
		grpcJsonVirtualHost.redirectRequest(outReq, req, true)
	}, grpcutil.NewTransportJson(grpcJsonVirtualHost.ClientCertificate))
}
