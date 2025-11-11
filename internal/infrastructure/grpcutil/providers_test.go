package grpcutil

import (
	"testing"

	certs "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/certificates"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewGrpcServer_WhenTransparentServer_ThenCreatesServerWithHandler(t *testing.T) {
	// Arrange
	grpcWebProxy := &GrpcWebProxy{
		GrpcProxy: GrpcProxy{
			IsTransparentServer: true,
		},
	}
	clientConn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials())) // This will fail but for test

	// Act
	server := NewGrpcServer(grpcWebProxy, clientConn)

	// Assert
	assert.NotNil(t, server)
}

func TestNewGrpcServer_WhenNotTransparentServer_ThenCreatesServerWithoutHandler(t *testing.T) {
	// Arrange
	grpcWebProxy := &GrpcWebProxy{
		GrpcProxy: GrpcProxy{
			IsTransparentServer: false,
		},
	}
	clientConn, _ := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))

	// Act
	server := NewGrpcServer(grpcWebProxy, clientConn)

	// Assert
	assert.NotNil(t, server)
}

func TestNewWrappedGrpcServer_WhenCalled_ThenReturnsWrappedServer(t *testing.T) {
	// Arrange
	grpcWebProxy := &GrpcWebProxy{
		AllowAllOrigins: true,
		UseWebSockets:   true,
	}
	grpcServer := grpc.NewServer()

	// Act
	wrapped := NewWrappedGrpcServer(grpcWebProxy, grpcServer)

	// Assert
	assert.NotNil(t, wrapped)
}

func TestGrpcWebProxy_makeHTTPOriginFunc_WhenAllowAllOrigins_ThenReturnsAllowingFunc(t *testing.T) {
	// Arrange
	proxy := &GrpcWebProxy{
		AllowAllOrigins: true,
	}

	// Act
	fn := proxy.makeHTTPOriginFunc()

	// Assert
	assert.True(t, fn("any-origin.com"))
}

func TestAllowedOrigins_toAllowedOriginsFormat_WhenCalled_ThenCreatesFormat(t *testing.T) {
	// Arrange
	origins := AllowedOrigins{"example.com", "test.com"}

	// Act
	format := origins.toAllowedOriginsFormat()

	// Assert
	assert.NotNil(t, format)
	assert.True(t, format.IsAllowed("example.com"))
	assert.False(t, format.IsAllowed("other.com"))
}

func TestNewGrpcClientConn_WhenNoCertificate_ThenUsesInsecure(t *testing.T) {
	// Arrange
	grpcWebProxy := &GrpcWebProxy{}
	var cert *certs.CertificateDefs

	// Act
	conn, err := NewGrpcClientConn(grpcWebProxy, cert, "invalid-host:1234")

	// Assert
	// gRPC Dial doesn't fail immediately for invalid hosts
	// It returns a connection that will fail on first use
	assert.NoError(t, err)
	assert.NotNil(t, conn)
}
