package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/janmbaco/go-reverseproxy-ssl/docker/grpc-server/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	hello.UnimplementedHelloServiceServer
}

func (s *server) SayHello(ctx context.Context, req *hello.HelloRequest) (*hello.HelloReply, error) {
	log.Printf("Received SayHello: %s", req.Name)
	return &hello.HelloReply{Message: "Hello " + req.Name}, nil
}

func (s *server) ServerStreamingChat(req *hello.HelloRequest, stream hello.HelloService_ServerStreamingChatServer) error {
	for i := 0; i < 3; i++ {
		if err := stream.Send(&hello.ChatMessage{Content: fmt.Sprintf("Server message %d for %s", i+1, req.Name)}); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (s *server) ClientStreamingChat(stream hello.HelloService_ClientStreamingChatServer) error {
	var messages []string
	for {
		in, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		messages = append(messages, in.Content)
	}
	return stream.SendAndClose(&hello.HelloReply{Message: fmt.Sprintf("Received: %v", messages)})
}

func (s *server) BidirectionalChat(stream hello.HelloService_BidirectionalChatServer) error {
	for {
		in, err := stream.Recv()
		if err != nil {
			return err
		}
		// Echo the message back
		if err := stream.Send(&hello.ChatMessage{Content: "Echo: " + in.Content}); err != nil {
			return err
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", ":7656")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	hello.RegisterHelloServiceServer(s, &server{})
	reflection.Register(s)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
