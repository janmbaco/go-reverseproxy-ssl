package grpcutil

import (
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

// WrappedGrpcServer is an HTTP handler that wraps a gRPC server to handle gRPC-Web requests.
// It converts incoming HTTP gRPC-Web requests to gRPC protocol and forwards them to the wrapped gRPC server.
type WrappedGrpcServer struct {
	wrappedServer *grpcweb.WrappedGrpcServer
}

// WrapServer creates a gRPC-Web HTTP handler that wraps a gRPC server.
// The returned handler can process HTTP requests with gRPC-Web protocol and forward them to the underlying gRPC server.
func WrapServer(server interface{}, options ...Option) *WrappedGrpcServer {
	grpcServer, ok := server.(*grpc.Server)
	if !ok {
		// Return stub if not a gRPC server
		return &WrappedGrpcServer{}
	}

	// Create basic gRPC-Web wrapper
	wrapped := grpcweb.WrapServer(grpcServer)
	wrappedGrpcServer := &WrappedGrpcServer{
		wrappedServer: wrapped,
	}

	// Apply options (for now, just ignore them as they're stub implementations)
	for _, opt := range options {
		opt(wrappedGrpcServer)
	}

	return wrappedGrpcServer
}

// Option is a stub option type
type Option func(*WrappedGrpcServer)

// WithCorsForRegisteredEndpointsOnly is a stub option
func WithCorsForRegisteredEndpointsOnly(corsForRegisteredOnly bool) Option {
	return func(w *WrappedGrpcServer) {}
}

// WithOriginFunc is a stub option
func WithOriginFunc(originFunc func(origin string) bool) Option {
	return func(w *WrappedGrpcServer) {}
}

// WithWebsockets is a stub option
func WithWebsockets(useWebsockets bool) Option {
	return func(w *WrappedGrpcServer) {}
}

// WithWebsocketOriginFunc is a stub option
func WithWebsocketOriginFunc(websocketOriginFunc func(req *http.Request) bool) Option {
	return func(w *WrappedGrpcServer) {}
}

// WithAllowedRequestHeaders is a stub option
func WithAllowedRequestHeaders(allowedHeaders []string) Option {
	return func(w *WrappedGrpcServer) {}
}

// ServeHTTP implements http.Handler for gRPC-Web requests
func (w *WrappedGrpcServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if w.wrappedServer == nil {
		http.Error(rw, "gRPC-Web not implemented", http.StatusNotImplemented)
		return
	}
	w.wrappedServer.ServeHTTP(rw, req)
}

// WebsocketRequestOrigin extracts the origin from a WebSocket request.
// This is a stub implementation that may be enhanced in the future.
func WebsocketRequestOrigin(req *http.Request) (string, error) {
	return "", nil
}
