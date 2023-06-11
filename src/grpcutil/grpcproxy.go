package grpcutil

const maxMsgSize = 1024 * 1024 * 4

// GrpcProxy is the object responsible to create a communication  gRPC Server that calls other gRPC Server.
type GrpcProxy struct {
	GrpcServices        map[string][]string `json:"grpc_services"`
	IsTransparentServer bool                `json:"is_transparent_server"`
	Authority           string              `json:"authority"`
}