package grpcUtil

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

type GrpcProxy struct {
	GrpcServices        map[string][]string `json:"grpc_services"`
	IsTransparentServer bool                `json:"is_transparent_server"`
	Authority           string              `json:"authority"`
}

func (this *GrpcProxy) NewServer(clientConn *grpc.ClientConn) *grpc.Server {

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

	if this.IsTransparentServer {
		grpcServer = grpc.NewServer(
			grpc.CustomCodec(proxy.Codec()), // needed for proxy to function.
			grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
			grpc.MaxRecvMsgSize(maxMsgSize))
	} else {
		grpcServer = grpc.NewServer(
			grpc.CustomCodec(proxy.Codec()), // needed for proxy to function.
			grpc.MaxRecvMsgSize(maxMsgSize))
	}

	if len(this.GrpcServices) > 0 {
		for serviceName, methodsNames := range this.GrpcServices {
			proxy.RegisterService(grpcServer, director, serviceName, methodsNames...)
		}
	}
	return grpcServer

}

func (this *GrpcProxy) CreateClientConn(clientCertificate *certs.CertificateDefs, host string) *grpc.ClientConn {
	var opt []grpc.DialOption
	opt = append(opt, grpc.WithCodec(proxy.Codec()))

	if len(this.Authority) > 0 {
		opt = append(opt, grpc.WithAuthority(this.Authority))
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
