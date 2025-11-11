package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestConfigUI_HTMLPages(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing Config UI HTML pages...")

	// Create HTTP client that skips TLS verification for local testing
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Test main config page
	resp, err := client.Get("http://localhost:8081/")
	if err != nil {
		t.Fatalf("Request to config UI failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	fmt.Printf("Config UI - Status: %s\n", resp.Status)
	fmt.Printf("Response length: %d bytes\n", len(body))

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for config UI, got %d", resp.StatusCode)
	}

	// Check for expected HTML content
	expectedElements := []string{
		"<title>",
		"Reverse Proxy Config",
		"Virtual Hosts",
		"SSL Certificates",
	}

	for _, element := range expectedElements {
		if !bytes.Contains(body, []byte(element)) {
			t.Errorf("Expected HTML to contain '%s', but it doesn't", element)
		}
	}

	fmt.Println("✓ Config UI HTML page loads correctly")
}

func TestConfigUI_VirtualHostsAPI(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing Config UI Virtual Hosts API...")

	client := createHTTPClient()

	// Test GET virtual hosts
	testGetVirtualHosts(t, client)

	// Test POST new virtual host
	createdID := testCreateVirtualHost(t, client)

	// Test DELETE virtual host
	if createdID != "" {
		testDeleteVirtualHost(t, client, createdID)
	}
}

func createHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func testGetVirtualHosts(t *testing.T, client *http.Client) {
	resp, err := client.Get("http://localhost:8081/api/virtualhosts")
	if err != nil {
		t.Fatalf("GET virtual hosts request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200 for GET virtual hosts, got %d", resp.StatusCode)
	}

	// Parse JSON response
	var virtualHosts map[string]interface{}
	if err := json.Unmarshal(body, &virtualHosts); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	fmt.Printf("Virtual hosts response: %s\n", string(body))
	fmt.Println("✓ GET virtual hosts API works correctly")
}

func testCreateVirtualHost(t *testing.T, client *http.Client) string {
	requestBody, contentType := createVirtualHostFormData(t)
	postReq := createPostRequest(t, requestBody, contentType)

	resp, err := client.Do(postReq)
	if err != nil {
		t.Fatalf("POST virtual host request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return handleVirtualHostCreationResponse(t, resp)
}

func createVirtualHostFormData(t *testing.T) (*bytes.Buffer, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	fields := map[string]string{
		"from":     "test.localhost",
		"scheme":   "http",
		"hostName": "backend-8080",
		"port":     "80",
	}

	for key, value := range fields {
		if err := w.WriteField(key, value); err != nil {
			t.Fatalf("Failed to write %s field: %v", key, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	return &b, w.FormDataContentType()
}

func createPostRequest(t *testing.T, body *bytes.Buffer, contentType string) *http.Request {
	postReq, err := http.NewRequest("POST", "http://localhost:8081/api/virtualhosts", body)
	if err != nil {
		t.Fatalf("Failed to create POST request: %v", err)
	}
	postReq.Header.Set("Content-Type", contentType)
	return postReq
}

func handleVirtualHostCreationResponse(t *testing.T, resp *http.Response) string {
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200/201 for POST virtual host, got %d. Response: %s", resp.StatusCode, string(body))
		return ""
	}

	fmt.Println("✓ POST virtual host API works correctly")
	return extractVirtualHostID(t, resp.Body)
}

func extractVirtualHostID(t *testing.T, body io.Reader) string {
	var response map[string]interface{}
	if err := json.NewDecoder(body).Decode(&response); err == nil {
		if id, ok := response["id"].(string); ok && id != "" {
			fmt.Printf("Created virtual host with ID: %s\n", id)
			return id
		}
	}
	return ""
}

func testDeleteVirtualHost(t *testing.T, client *http.Client, virtualHostID string) {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("http://localhost:8081/api/virtualhosts/%s", virtualHostID), nil)
	if err != nil {
		t.Fatalf("Failed to create DELETE request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("DELETE virtual host request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200/204 for DELETE virtual host, got %d. Response: %s", resp.StatusCode, string(body))
	} else {
		fmt.Println("✓ DELETE virtual host API works correctly")
	}
}

func TestConfigUI_CertificatesAPI(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing Config UI Certificates API...")

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Create a simple test certificate file content
	certContent := `-----BEGIN CERTIFICATE-----
MIICiTCCAg+gAwIBAgIJAJ8l2Z2Z3Z3ZMAOGA1UEBhMCVVMxCzAJBgNVBAgTAkNB
MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRowGAYDVQQKExFPcGVuU1NMIENlcnRp
ZmljYXRlIEF1dGhvcml0eTELMAkGA1UECxMCSVQxFjAUBgNVBAMTDU9wZW5TU0wg
Q0EgQ2VydDAeFw0xNTA0MTUxNzI5NTlaFw0yNTA0MTIxNzI5NTlaMB4xHDAaBgNV
BAMTE3d3dy5leGFtcGxlLmNvbXBhbnkwWjANBgkqhkiG9w0BAQEFAANOCQDNAQAB
...
-----END CERTIFICATE-----`

	keyContent := `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC7VJTUt9Us8cKB
...
-----END PRIVATE KEY-----`

	// Test certificate upload
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Add certificate file
	certWriter, err := w.CreateFormFile("certificate", "test-cert.pem")
	if err != nil {
		t.Fatalf("Failed to create form file for certificate: %v", err)
	}
	_, err = certWriter.Write([]byte(certContent))
	if err != nil {
		t.Fatalf("Failed to write certificate content: %v", err)
	}

	// Add key file
	keyWriter, err := w.CreateFormFile("key", "test-key.pem")
	if err != nil {
		t.Fatalf("Failed to create form file for key: %v", err)
	}
	_, err = keyWriter.Write([]byte(keyContent))
	if err != nil {
		t.Fatalf("Failed to write key content: %v", err)
	}

	// Add domain name
	err = w.WriteField("domain", "test.example.com")
	if err != nil {
		t.Fatalf("Failed to write domain field: %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("Failed to close multipart writer: %v", err)
	}

	req, err := http.NewRequest("POST", "http://localhost:8081/api/certificates", &b)
	if err != nil {
		t.Fatalf("Failed to create certificate upload request: %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Certificate upload request failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read certificate upload response: %v", err)
	}

	fmt.Printf("Certificate upload - Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))

	// Certificate upload might fail due to invalid test certificates, but the endpoint should respond
	if resp.StatusCode >= 500 {
		t.Errorf("Certificate upload endpoint returned server error: %d", resp.StatusCode)
	} else {
		fmt.Println("✓ Certificate upload API endpoint responds correctly")
	}
}

func TestConfigUI_ServerStatus(t *testing.T) {
	// Skip test if INTEGRATION_TEST environment variable is not set
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Skipping integration test. Set INTEGRATION_TEST=true to run.")
	}

	fmt.Println("Testing Config UI server status...")

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// Test health check endpoint
	resp, err := client.Get("http://localhost:8081/health")
	if err != nil {
		t.Logf("Health check endpoint not available (expected): %v", err)
	} else {
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Health check - Status: %d, Response: %s\n", resp.StatusCode, string(body))
	}

	// Test that the server is responding to basic requests
	resp, err = client.Get("http://localhost:8081/")
	if err != nil {
		t.Fatalf("Config UI server is not responding: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Errorf("Config UI server returned status %d, expected 200", resp.StatusCode)
	} else {
		fmt.Println("✓ Config UI server is running and responding correctly")
	}
}
