package grpcUtil

import (
	"context"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
)

type GrpcWebProxy struct {
	AllowAllOrigins      bool           `json:"allow_all_origins"`
	AllowedOrigins       AllowedOrigins `json:"allowed_origins"`
	UseWebSockets        bool           `json:"use_web_sockets"`
	AllowedHeaders       []string       `json:"allowed_header"`
	allowedOriginsFormat *allowedOriginsFormat
}

func (this *GrpcWebProxy) WrappedGrpcServer(clientConn *grpc.ClientConn) *grpcweb.WrappedGrpcServer {

	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		outCtx, _ := context.WithCancel(ctx)
		mdCopy := md.Copy()
		delete(mdCopy, "user-agent")
		delete(mdCopy, "connection")
		outCtx = metadata.NewOutgoingContext(outCtx, mdCopy)
		return outCtx, clientConn, nil
	}

	grpcServer := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()), // needed for proxy to function.
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
		grpc.MaxRecvMsgSize(1024*1024*4))

	if this.AllowedOrigins == nil {
		this.AllowedOrigins = make([]string, 0)
	}

	this.allowedOriginsFormat = this.AllowedOrigins.toAllowedOriginsFormat()

	options := []grpcweb.Option{
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(this.makeHttpOriginFunc()),
	}

	if this.UseWebSockets {
		options = append(
			options,
			grpcweb.WithWebsockets(true),
			grpcweb.WithWebsocketOriginFunc(this.getWebsocketOriginFunc()),
		)
	}

	if this.AllowedHeaders != nil && len(this.AllowedHeaders) > 0 {
		options = append(
			options,
			grpcweb.WithAllowedRequestHeaders(this.AllowedHeaders),
		)
	}

	return grpcweb.WrapServer(grpcServer, options...)

}

func (this *GrpcWebProxy) makeHttpOriginFunc() func(origin string) bool {
	if this.AllowAllOrigins {
		return func(origin string) bool {
			return true
		}
	}
	return this.allowedOriginsFormat.IsAllowed
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

func (this *GrpcWebProxy) getWebsocketOriginFunc() func(req *http.Request) bool {
	var result func(req *http.Request) bool
	if this.AllowAllOrigins {
		result = func(req *http.Request) bool {
			return true
		}
	} else {
		result = this.allowedOriginsFormat.getWebsocketOriginFunc()
	}
	return result
}

type AllowedOrigins []string

type allowedOriginsFormat struct {
	origins map[string]struct{}
}

func (allowedOrigins *allowedOriginsFormat) getWebsocketOriginFunc() func(req *http.Request) bool {
	return func(req *http.Request) bool {
		origin, err := grpcweb.WebsocketRequestOrigin(req)
		if err != nil {
			grpclog.Warning(err)
			return false
		}
		return allowedOrigins.IsAllowed(origin)
	}
}

func (a *allowedOriginsFormat) IsAllowed(origin string) bool {
	_, ok := a.origins[origin]
	return ok
}
