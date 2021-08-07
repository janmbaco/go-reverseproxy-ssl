package grpcutil

import (
	"context"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

// NewGrpcServer returns a grpc Server
func NewGrpcServer(grpcProxy *GrpcProxy, clientConn *grpc.ClientConn) *grpc.Server {
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		outCtx, _ := context.WithCancel(ctx) //nolint:govet
		mdCopy := md.Copy()
		delete(mdCopy, "user-agent")
		delete(mdCopy, "connection")
		outCtx = metadata.NewOutgoingContext(outCtx, mdCopy)
		return outCtx, clientConn, nil
	}
	var grpcServer *grpc.Server

	if grpcProxy.IsTransparentServer {
		grpcServer = grpc.NewServer(
			grpc.CustomCodec(proxy.Codec()), // needed for proxy to function.
			grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
			grpc.MaxRecvMsgSize(maxMsgSize))
	} else {
		grpcServer = grpc.NewServer(
			grpc.CustomCodec(proxy.Codec()), // needed for proxy to function.
			grpc.MaxRecvMsgSize(maxMsgSize))
	}

	if len(grpcProxy.GrpcServices) > 0 {
		for serviceName, methodsNames := range grpcProxy.GrpcServices {
			proxy.RegisterService(grpcServer, director, serviceName, methodsNames...)
		}
	}
	return grpcServer
}

// NewWrappedGrpcServer returns a gRPC Web wrapped server.
func NewWrappedGrpcServer(grpcWebProxy *GrpcWebProxy, grpcServer *grpc.Server) *grpcweb.WrappedGrpcServer {
	if grpcWebProxy.AllowedOrigins == nil {
		grpcWebProxy.AllowedOrigins = make([]string, 0)
	}
	grpcWebProxy.allowedOriginsFormat = grpcWebProxy.AllowedOrigins.toAllowedOriginsFormat()
	options := []grpcweb.Option{
		grpcweb.WithCorsForRegisteredEndpointsOnly(false),
		grpcweb.WithOriginFunc(grpcWebProxy.makeHTTPOriginFunc()),
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

// NewGrpcClientConn return a client for a gRPC server.
func NewGrpcClientConn(grpcProxy *GrpcProxy, clientCertificate *certs.CertificateDefs, hostName string) *grpc.ClientConn {
	var opt []grpc.DialOption
	opt = append(opt, grpc.WithCodec(proxy.Codec()))

	if len(grpcProxy.Authority) > 0 {
		opt = append(opt, grpc.WithAuthority(grpcProxy.Authority))
	}

	if clientCertificate != nil {
		opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(clientCertificate.GetTLSConfig())))
	} else {
		opt = append(opt, grpc.WithInsecure())
	}

	clientConn, err := grpc.Dial(hostName, opt...)
	errorschecker.TryPanic(err)

	return clientConn
}
