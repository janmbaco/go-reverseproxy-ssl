package grpcutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestWrapServer_WhenValidGrpcServer_ThenReturnsWrappedServer(t *testing.T) {
	// Arrange
	grpcServer := grpc.NewServer()

	// Act
	wrapped := WrapServer(grpcServer)

	// Assert
	assert.NotNil(t, wrapped)
	assert.NotNil(t, wrapped.wrappedServer)
}

func TestWrapServer_WhenInvalidServer_ThenReturnsStub(t *testing.T) {
	// Arrange
	invalidServer := "not a grpc server"

	// Act
	wrapped := WrapServer(invalidServer)

	// Assert
	assert.NotNil(t, wrapped)
	assert.Nil(t, wrapped.wrappedServer)
}

func TestWrappedGrpcServer_ServeHTTP_WhenWrappedServerNil_ThenReturnsNotImplemented(t *testing.T) {
	// Arrange
	wrapped := &WrappedGrpcServer{wrappedServer: nil}
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	// Act
	wrapped.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotImplemented, w.Code)
	assert.Contains(t, w.Body.String(), "gRPC-Web not implemented")
}

func TestWrappedGrpcServer_ServeHTTP_WhenValidWrappedServer_ThenDelegates(t *testing.T) {
	// Arrange
	grpcServer := grpc.NewServer()
	wrapped := WrapServer(grpcServer)
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	// Act
	wrapped.ServeHTTP(w, req)

	// Assert
	// grpcweb.WrappedGrpcServer should handle the request
	// Since we don't have a real gRPC-Web request, it may return an error
	// but the important thing is that it doesn't panic and delegates properly
	assert.NotNil(t, wrapped.wrappedServer)
}

func TestWithCorsForRegisteredEndpointsOnly_WhenCalled_ThenReturnsOption(t *testing.T) {
	// Act
	option := WithCorsForRegisteredEndpointsOnly(true)

	// Assert
	assert.NotNil(t, option)
}

func TestWithOriginFunc_WhenCalled_ThenReturnsOption(t *testing.T) {
	// Arrange
	originFunc := func(origin string) bool { return true }

	// Act
	option := WithOriginFunc(originFunc)

	// Assert
	assert.NotNil(t, option)
}

func TestWithWebsockets_WhenCalled_ThenReturnsOption(t *testing.T) {
	// Act
	option := WithWebsockets(true)

	// Assert
	assert.NotNil(t, option)
}

func TestWithWebsocketOriginFunc_WhenCalled_ThenReturnsOption(t *testing.T) {
	// Arrange
	wsFunc := func(req *http.Request) bool { return true }

	// Act
	option := WithWebsocketOriginFunc(wsFunc)

	// Assert
	assert.NotNil(t, option)
}

func TestWithAllowedRequestHeaders_WhenCalled_ThenReturnsOption(t *testing.T) {
	// Arrange
	headers := []string{"content-type", "x-grpc-web"}

	// Act
	option := WithAllowedRequestHeaders(headers)

	// Assert
	assert.NotNil(t, option)
}

func TestWebsocketRequestOrigin_WhenCalled_ThenReturnsEmptyString(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "/test", nil)

	// Act
	origin, err := WebsocketRequestOrigin(req)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "", origin)
}
