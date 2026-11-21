package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestHttpThroughProxy(t *testing.T) {
	client := integrationClient(t)
	// The Compose fixture generates a new self-signed localhost certificate on each run.
	client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	defer client.CloseIdleConnections()
	for i, path := range []string{"/", "/prueba-1/", "/prueba-2/", "/prueba-3/"} {
		resp, err := client.Get("https://localhost" + path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		want := fmt.Sprintf("Service on port %d", 8080+i)
		if err != nil || resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != want {
			t.Errorf("GET %s: status=%d body=%q, want %q, read error=%v", path, resp.StatusCode, body, want, err)
		}
	}
}
