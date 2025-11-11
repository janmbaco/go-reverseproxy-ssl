package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

type HelloRequest struct {
	Name string
}

type HelloReply struct {
	Message string `protobuf:"bytes,1,opt,name=message,proto3"`
}

func TestGrpcWebThroughProxy(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing gRPC-Web through reverse proxy...")

	// gRPC-Web request payload
	name := "Integration Test"
	nameBytes := []byte(name)
	data := append([]byte{0x0A, byte(len(nameBytes))}, nameBytes...) // tag 1, length, string

	// Frame the data: 1 byte type (0 for unary), 4 bytes length, then data
	frame := make([]byte, 5+len(data))
	frame[0] = 0 // unary request
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)

	// Create HTTP request to gRPC-Web endpoint
	httpReq, err := http.NewRequest("POST", "https://localhost/grpc/hello.HelloService/SayHello", bytes.NewBuffer(frame))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set gRPC-Web headers
	httpReq.Header.Set("Content-Type", "application/grpc-web")
	httpReq.Header.Set("X-Grpc-Web", "1")

	// Use TLS with skip verify for local testing
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check response contains expected message
	expected := "Hello Integration Test"
	if !bytes.Contains(body, []byte(expected)) {
		t.Errorf("Expected response to contain '%s', got %s", expected, string(body))
	}
}

func TestGrpcWebBidirectionalThroughProxy(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing gRPC-Web bidirectional streaming through reverse proxy...")

	// gRPC-Web bidirectional chat: send a message and expect echo
	message := "Test bidirectional"
	messageBytes := []byte(message)
	data := append([]byte{0x0A, byte(len(messageBytes))}, messageBytes...) // tag 1, length, string

	// Frame: 1 byte type (0 for data), 4 bytes length, then data
	frame := make([]byte, 5+len(data))
	frame[0] = 0 // data frame
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)

	// Create HTTP request to gRPC-Web endpoint
	httpReq, err := http.NewRequest("POST", "https://localhost/grpc/hello.HelloService/BidirectionalChat", bytes.NewBuffer(frame))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set gRPC-Web headers
	httpReq.Header.Set("Content-Type", "application/grpc-web")
	httpReq.Header.Set("X-Grpc-Web", "1")

	// Use TLS with skip verify
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	_ = http2.ConfigureTransport(client.Transport.(*http.Transport))

	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response length: %d\n", len(body))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check if response contains "Echo: Test bidirectional"
	expected := "Echo: Test bidirectional"
	if !bytes.Contains(body, []byte(expected)) {
		t.Errorf("Expected response to contain '%s', got %s", expected, string(body))
	}
}

func TestGrpcWebServerStreamingThroughProxy(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing gRPC-Web server streaming through reverse proxy...")

	// Similar to unary but expect multiple messages
	name := "Server Streaming Test"
	nameBytes := []byte(name)
	data := append([]byte{0x0A, byte(len(nameBytes))}, nameBytes...)

	frame := make([]byte, 5+len(data))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)

	httpReq, err := http.NewRequest("POST", "https://localhost/grpc/hello.HelloService/ServerStreamingChat", bytes.NewBuffer(frame))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/grpc-web")
	httpReq.Header.Set("X-Grpc-Web", "1")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response length: %d\n", len(body))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check for server messages
	if !bytes.Contains(body, []byte("Server message")) {
		t.Errorf("Expected server streaming messages, got %s", string(body))
	}
}

func TestGrpcWebClientStreamingThroughProxy(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing gRPC-Web client streaming through reverse proxy...")

	// Send multiple messages
	messages := []string{"Client msg 1", "Client msg 2"}
	var frames []byte
	for _, msg := range messages {
		msgBytes := []byte(msg)
		data := append([]byte{0x0A, byte(len(msgBytes))}, msgBytes...)
		frame := make([]byte, 5+len(data))
		frame[0] = 0
		binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
		copy(frame[5:], data)
		frames = append(frames, frame...)
	}

	httpReq, err := http.NewRequest("POST", "https://localhost/grpc/hello.HelloService/ClientStreamingChat", bytes.NewBuffer(frames))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/grpc-web")
	httpReq.Header.Set("X-Grpc-Web", "1")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response length: %d\n", len(body))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check for received messages
	if !bytes.Contains(body, []byte("Received:")) {
		t.Errorf("Expected client streaming response, got %s", string(body))
	}
}

// Test selective gRPC-Web proxy (only specific services/methods)
func TestGrpcWebSelectiveProxy(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing gRPC-Web selective proxy (allowed method)...")

	// Test allowed method: SayHello
	name := "Selective Test"
	nameBytes := []byte(name)
	data := append([]byte{0x0A, byte(len(nameBytes))}, nameBytes...)
	frame := make([]byte, 5+len(data))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)

	httpReq, err := http.NewRequest("POST", "https://localhost/grpc-selective/hello.HelloService/SayHello", bytes.NewBuffer(frame))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/grpc-web")
	httpReq.Header.Set("X-Grpc-Web", "1")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for allowed method, got %d", resp.StatusCode)
	}

	if !bytes.Contains(body, []byte("Hello")) {
		t.Errorf("Expected greeting response, got %s", string(body))
	}
}

func TestGrpcWebSelectiveProxyDisallowedMethod(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing gRPC-Web selective proxy (disallowed method)...")

	// Test disallowed method: BidirectionalChat (not in grpc_services list)
	name := "Disallowed Test"
	nameBytes := []byte(name)
	data := append([]byte{0x0A, byte(len(nameBytes))}, nameBytes...)
	frame := make([]byte, 5+len(data))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)

	httpReq, err := http.NewRequest("POST", "https://localhost/grpc-selective/hello.HelloService/BidirectionalChat", bytes.NewBuffer(frame))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	httpReq.Header.Set("Content-Type", "application/grpc-web")
	httpReq.Header.Set("X-Grpc-Web", "1")

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Response length: %d\n", len(body))
	fmt.Printf("Response body (hex): %x\n", body)
	fmt.Printf("Response body (string): %q\n", string(body))

	// In gRPC-Web, errors are encoded in the response body with grpc-status trailer
	// A 200 OK with empty body or grpc-status != 0 indicates an error
	if resp.StatusCode != 200 {
		fmt.Printf("✓ Correctly rejected with HTTP status %d\n", resp.StatusCode)
	} else if len(body) == 0 {
		fmt.Println("✓ Correctly rejected with empty response (likely Unimplemented)")
	} else if bytes.Contains(body, []byte("grpc-status: 12")) || bytes.Contains(body, []byte("grpc-status:12")) {
		fmt.Println("✓ Correctly rejected with grpc-status 12 (Unimplemented)")
	} else {
		t.Errorf("Expected error for disallowed method, got successful response")
	}
}

func TestGrpcWebTransparentVsSelective(t *testing.T) {
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing transparent vs selective gRPC-Web proxy comparison...")

	name := "Comparison Test"
	nameBytes := []byte(name)
	data := append([]byte{0x0A, byte(len(nameBytes))}, nameBytes...)
	frame := make([]byte, 5+len(data))
	frame[0] = 0
	binary.BigEndian.PutUint32(frame[1:5], uint32(len(data)))
	copy(frame[5:], data)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Test BidirectionalChat on transparent proxy (should work)
	reqTransparent, _ := http.NewRequest("POST", "https://localhost/grpc/hello.HelloService/BidirectionalChat", bytes.NewBuffer(frame))
	reqTransparent.Header.Set("Content-Type", "application/grpc-web")
	reqTransparent.Header.Set("X-Grpc-Web", "1")

	respTransparent, err := client.Do(reqTransparent)
	if err != nil {
		t.Fatalf("Transparent proxy request failed: %v", err)
	}
	defer func() { _ = respTransparent.Body.Close() }()

	// Test BidirectionalChat on selective proxy (should fail - not in allowed list)
	reqSelective, _ := http.NewRequest("POST", "https://localhost/grpc-selective/hello.HelloService/BidirectionalChat", bytes.NewBuffer(frame))
	reqSelective.Header.Set("Content-Type", "application/grpc-web")
	reqSelective.Header.Set("X-Grpc-Web", "1")

	respSelective, err := client.Do(reqSelective)
	if err != nil {
		t.Fatalf("Selective proxy request failed: %v", err)
	}
	defer func() { _ = respSelective.Body.Close() }()

	bodyTransparent, _ := io.ReadAll(respTransparent.Body)
	bodySelective, _ := io.ReadAll(respSelective.Body)

	fmt.Printf("Transparent proxy status: %d, body length: %d\n", respTransparent.StatusCode, len(bodyTransparent))
	fmt.Printf("Selective proxy status: %d, body length: %d\n", respSelective.StatusCode, len(bodySelective))

	if respTransparent.StatusCode != 200 || len(bodyTransparent) == 0 {
		t.Errorf("Transparent proxy should allow BidirectionalChat, got status %d with body length %d", respTransparent.StatusCode, len(bodyTransparent))
	}

	// Selective proxy should reject (empty body or error status)
	if respSelective.StatusCode == 200 && len(bodySelective) > 0 && !bytes.Contains(bodySelective, []byte("grpc-status: 12")) {
		t.Errorf("Selective proxy should reject BidirectionalChat (not in allowed list), got status %d with body length %d", respSelective.StatusCode, len(bodySelective))
	} else {
		fmt.Println("✓ Transparent proxy allows all methods, selective proxy restricts correctly")
	}
}
