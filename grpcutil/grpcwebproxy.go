package grpcutil

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// GrpcWebProxy is the object responsible to create a communication  gRPC Web Server that calls other gRPC Server.
type GrpcWebProxy struct {
	GrpcProxy
	AllowAllOrigins      bool           `json:"allow_all_origins"`
	AllowedOrigins       AllowedOrigins `json:"allowed_origins"`
	UseWebSockets        bool           `json:"use_web_sockets"`
	AllowedHeaders       []string       `json:"allowed_header"`
	allowedOriginsFormat *allowedOriginsFormat
}

// WrappedGrpcServer returns a gRPC Web wrapped server.
func (grpcWebProxy *GrpcWebProxy) WrappedGrpcServer(clientConn *grpc.ClientConn) *grpcweb.WrappedGrpcServer {

	grpcServer := grpcWebProxy.NewServer(clientConn)

	if grpcWebProxy.AllowedOrigins == nil {
		grpcWebProxy.AllowedOrigins = make([]string, 0)
	}

	grpcWebProxy.allowedOriginsFormat = grpcWebProxy.AllowedOrigins.toAllowedOriginsFormat()

	options := []grpcweb.Option{
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(grpcWebProxy.makeHttpOriginFunc()),
	}

	if grpcWebProxy.UseWebSockets {
		options = append(
			options,
			grpcweb.WithWebsockets(true),
			grpcweb.WithWebsocketOriginFunc(grpcWebProxy.getWebsocketOriginFunc()),
		)
	}

	if grpcWebProxy.AllowedHeaders != nil && len(grpcWebProxy.AllowedHeaders) > 0 {
		options = append(
			options,
			grpcweb.WithAllowedRequestHeaders(grpcWebProxy.AllowedHeaders),
		)
	}

	return grpcweb.WrapServer(grpcServer, options...)

}

func (grpcWebProxy *GrpcWebProxy) makeHttpOriginFunc() func(origin string) bool {
	if grpcWebProxy.AllowAllOrigins {
		return func(origin string) bool {
			return true
		}
	}
	return grpcWebProxy.allowedOriginsFormat.IsAllowed
}

func (origins AllowedOrigins) toAllowedOriginsFormat() *allowedOriginsFormat {
	o := map[string]struct{}{}
	for _, allowedOrigin := range origins {
		o[allowedOrigin] = struct{}{}
	}
	return &allowedOriginsFormat{
		origins: o,
	}
}

func (grpcWebProxy *GrpcWebProxy) getWebsocketOriginFunc() func(req *http.Request) bool {
	var result func(req *http.Request) bool
	if grpcWebProxy.AllowAllOrigins {
		result = func(req *http.Request) bool {
			return true
		}
	} else {
		result = grpcWebProxy.allowedOriginsFormat.getWebsocketOriginFunc()
	}
	return result
}

// AllowedOrigins is used to register de Allowed Origins.
type AllowedOrigins []string

type allowedOriginsFormat struct {
	origins map[string]struct{}
}

func (allowedOriginsFormat *allowedOriginsFormat) getWebsocketOriginFunc() func(req *http.Request) bool {
	return func(req *http.Request) bool {
		origin, err := grpcweb.WebsocketRequestOrigin(req)
		if err != nil {
			grpclog.Warning(err)
			return false
		}
		return allowedOriginsFormat.IsAllowed(origin)
	}
}

func (allowedOriginsFormat *allowedOriginsFormat) IsAllowed(origin string) bool {
	_, ok := allowedOriginsFormat.origins[origin]
	return ok
}
