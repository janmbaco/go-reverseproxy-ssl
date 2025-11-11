package grpcutil

import (
	"net/http"
)

// WrappedGrpcServer is a local stub implementation that replaces grpc-web functionality
type WrappedGrpcServer struct{}

// WrapServer creates a stub wrapped server
func WrapServer(server interface{}, options ...Option) *WrappedGrpcServer {
	return &WrappedGrpcServer{}
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

// ServeHTTP is a stub implementation
func (w *WrappedGrpcServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	http.Error(rw, "gRPC-Web not implemented", http.StatusNotImplemented)
}

// WebsocketRequestOrigin is a stub function
func WebsocketRequestOrigin(req *http.Request) (string, error) {
	return "", nil
}
