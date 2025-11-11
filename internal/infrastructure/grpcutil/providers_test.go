package grpcutil

import (
	"net/http/httptest"
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

func TestGrpcWebProxy_makeHTTPOriginFunc_WhenNotAllowAllOrigins_ThenReturnsSelectiveFunc(t *testing.T) {
	// Arrange
	proxy := &GrpcWebProxy{
		AllowAllOrigins: false,
		AllowedOrigins:  AllowedOrigins{"example.com"},
	}
	proxy.allowedOriginsFormat = proxy.AllowedOrigins.toAllowedOriginsFormat()

	// Act
	fn := proxy.makeHTTPOriginFunc()

	// Assert
	assert.True(t, fn("example.com"))
	assert.False(t, fn("other.com"))
}

func TestGrpcWebProxy_getWebsocketOriginFunc_WhenAllowAllOrigins_ThenReturnsAllowingFunc(t *testing.T) {
	// Arrange
	proxy := &GrpcWebProxy{
		AllowAllOrigins: true,
	}

	// Act
	fn := proxy.getWebsocketOriginFunc()

	// Assert
	req := httptest.NewRequest("GET", "/test", nil)
	assert.True(t, fn(req))
}

func TestGrpcWebProxy_getWebsocketOriginFunc_WhenNotAllowAllOrigins_ThenReturnsSelectiveFunc(t *testing.T) {
	// Arrange
	proxy := &GrpcWebProxy{
		AllowAllOrigins: false,
		AllowedOrigins:  AllowedOrigins{"example.com"},
	}
	proxy.allowedOriginsFormat = proxy.AllowedOrigins.toAllowedOriginsFormat()

	// Act
	fn := proxy.getWebsocketOriginFunc()

	// Assert
	req := httptest.NewRequest("GET", "/test", nil)
	// Since WebsocketRequestOrigin returns empty string, and empty string is not in allowed origins
	assert.False(t, fn(req))
}

func TestAllowedOriginsFormat_getWebsocketOriginFunc_WhenCalled_ThenReturnsFunc(t *testing.T) {
	// Arrange
	format := &allowedOriginsFormat{
		origins: map[string]struct{}{
			"example.com": {},
		},
	}

	// Act
	fn := format.getWebsocketOriginFunc()

	// Assert
	req := httptest.NewRequest("GET", "/test", nil)
	// Since WebsocketRequestOrigin returns empty string, it should return false
	assert.False(t, fn(req))
}

func TestAllowedOriginsFormat_IsAllowed_WhenOriginAllowed_ThenReturnsTrue(t *testing.T) {
	// Arrange
	format := &allowedOriginsFormat{
		origins: map[string]struct{}{
			"example.com": {},
		},
	}

	// Act & Assert
	assert.True(t, format.IsAllowed("example.com"))
	assert.False(t, format.IsAllowed("other.com"))
}
