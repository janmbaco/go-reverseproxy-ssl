package grpcutil

import (
	"context"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	certs "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/certificates"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// NewGrpcServer returns a grpc Server
func NewGrpcServer(grpcWebProxy *GrpcWebProxy, clientConn *grpc.ClientConn) *grpc.Server {
	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		mdCopy := md.Copy()
		delete(mdCopy, "user-agent")
		delete(mdCopy, "connection")
		outCtx := metadata.NewOutgoingContext(ctx, mdCopy)
		return outCtx, clientConn, nil
	}
	var grpcServer *grpc.Server

	if grpcWebProxy.IsTransparentServer {
		// Transparent mode: proxy ALL services and methods
		grpcServer = grpc.NewServer(
			grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
			grpc.MaxRecvMsgSize(maxMsgSize))
	} else {
		// Selective mode: only proxy explicitly registered services/methods
		grpcServer = grpc.NewServer(
			grpc.MaxRecvMsgSize(maxMsgSize))

		// Register only the specified services and methods
		if len(grpcWebProxy.GrpcServices) > 0 {
			for serviceName, methodsNames := range grpcWebProxy.GrpcServices {
				proxy.RegisterService(grpcServer, director, serviceName, methodsNames...)
			}
		}
		// Note: When no services are registered and no UnknownServiceHandler is set,
		// all requests will return grpc "Unimplemented" error
	}

	return grpcServer
}

// NewWrappedGrpcServer returns a gRPC Web wrapped server.
func NewWrappedGrpcServer(grpcWebProxy *GrpcWebProxy, grpcServer *grpc.Server) *grpcweb.WrappedGrpcServer {
	if grpcWebProxy.AllowedOrigins == nil {
		grpcWebProxy.AllowedOrigins = make([]string, 0)
	}
	grpcWebProxy.allowedOriginsFormat = grpcWebProxy.AllowedOrigins.toAllowedOriginsFormat()

	// When IsTransparentServer is false, only allow CORS for registered endpoints
	// This ensures that only explicitly registered services/methods are accessible
	corsForRegisteredOnly := !grpcWebProxy.IsTransparentServer

	options := []grpcweb.Option{
		grpcweb.WithCorsForRegisteredEndpointsOnly(corsForRegisteredOnly),
		grpcweb.WithOriginFunc(grpcWebProxy.makeHTTPOriginFunc()),
	}
	if grpcWebProxy.UseWebSockets {
		options = append(
			options,
			grpcweb.WithWebsockets(true),
			grpcweb.WithWebsocketOriginFunc(grpcWebProxy.getWebsocketOriginFunc()),
		)
	}
	if len(grpcWebProxy.AllowedHeaders) > 0 {
		options = append(
			options,
			grpcweb.WithAllowedRequestHeaders(grpcWebProxy.AllowedHeaders),
		)
	}
	return grpcweb.WrapServer(grpcServer, options...)
}

// NewGrpcClientConn return a client for a gRPC server.
func NewGrpcClientConn(grpcWebProxy *GrpcWebProxy, clientCertificate *certs.CertificateDefs, hostName string) (*grpc.ClientConn, error) {
	var opt []grpc.DialOption

	if len(grpcWebProxy.Authority) > 0 {
		opt = append(opt, grpc.WithAuthority(grpcWebProxy.Authority))
	}

	if clientCertificate != nil {
		tlsConfig, err := clientCertificate.GetTLSConfig()
		if err != nil {
			return nil, err
		}
		opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		opt = append(opt, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	clientConn, err := grpc.NewClient(hostName, opt...)
	if err != nil {
		return nil, err
	}

	return clientConn, nil
}
