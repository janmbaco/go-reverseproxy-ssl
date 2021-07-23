package grpcutil

import (
	"context"

	"github.com/janmbaco/go-infrastructure/errorhandler"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const maxMsgSize = 1024 * 1024 * 4

// GrpcProxy is the object responsible to create a communication  gRPC Server that calls other gRPC Server.
type GrpcProxy struct {
	GrpcServices        map[string][]string `json:"grpc_services"`
	IsTransparentServer bool                `json:"is_transparent_server"`
	Authority           string              `json:"authority"`
}

// NewServer returns a grpc Server
func (grpcProxy *GrpcProxy) NewServer(clientConn *grpc.ClientConn) *grpc.Server {

	director := func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		outCtx, _ := context.WithCancel(ctx)
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

// CreateClientConn return a Client for a gRPC server.
func (grpcProxy *GrpcProxy) CreateClientConn(clientCertificate *certs.CertificateDefs, host string) *grpc.ClientConn {
	var opt []grpc.DialOption
	opt = append(opt, grpc.WithCodec(proxy.Codec()))

	if len(grpcProxy.Authority) > 0 {
		opt = append(opt, grpc.WithAuthority(grpcProxy.Authority))
	}

	if clientCertificate != nil {
		opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(clientCertificate.GetTlsConfig())))
	} else {
		opt = append(opt, grpc.WithInsecure())
	}

	clientConn, err := grpc.Dial(host, opt...)
	errorhandler.TryPanic(err)

	return clientConn
}
