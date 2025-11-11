package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestHttpThroughProxy(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing HTTP through reverse proxy...")

	// Use TLS with skip verify for local testing
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Test main host
	resp, err := client.Get("https://localhost/")
	if err != nil {
		t.Fatalf("Request to localhost failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Main host - Status: %s\n", resp.Status)
	fmt.Printf("Response length: %d bytes\n", len(body))
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for main host, got %d", resp.StatusCode)
	}

	// Test sub-path
	resp2, err := client.Get("https://localhost/prueba-1/")
	if err != nil {
		t.Fatalf("Request to localhost/prueba-1 failed: %v", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Sub-path - Status: %s\n", resp2.Status)
	fmt.Printf("Response length: %d bytes\n", len(body2))
	if resp2.StatusCode != 200 {
		t.Errorf("Expected status 200 for sub-path, got %d", resp2.StatusCode)
	}
}
