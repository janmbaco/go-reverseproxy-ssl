package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

const configUI = "http://localhost:8081"

func integrationClient(t *testing.T) *http.Client {
	t.Helper()
	if os.Getenv("INTEGRATION_TEST") != "true" {
		t.Skip("Set INTEGRATION_TEST=true with the integration stack running.")
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func TestConfigUI_HTMLPages(t *testing.T) {
	client := integrationClient(t)
	for _, page := range []struct{ path, title string }{
		{"/", "Dashboard - Reverse Proxy Config"},
		{"/virtualhosts", "Virtual Hosts - Reverse Proxy Config"},
		{"/config", "Configuration - Reverse Proxy Config"},
	} {
		res, err := client.Get(configUI + page.path)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if err != nil || res.StatusCode != http.StatusOK || !bytes.Contains(body, []byte(page.title)) {
			t.Fatalf("GET %s: status=%d, expected page title %q, read error=%v", page.path, res.StatusCode, page.title, err)
		}
	}
}

func TestConfigUI_ConfigurationAPI(t *testing.T) {
	client := integrationClient(t)
	res, err := client.Get(configUI + "/api/config")
	if err != nil {
		t.Fatal(err)
	}
	var config struct {
		DefaultHost string `json:"default_host"`
		ConfigPort  string `json:"config_ui_port"`
	}
	decodeOK(t, res, &config)
	if config.DefaultHost != "localhost" || config.ConfigPort != ":8081" {
		t.Fatalf("unexpected active fixture: host=%q config port=%q", config.DefaultHost, config.ConfigPort)
	}
}

func TestConfigUI_VirtualHostsAPI(t *testing.T) {
	client := integrationClient(t)
	const host = "route-test.localhost"
	id := createHost(t, client, host, nil)
	if got := hostID(t, client, host); got != id {
		t.Fatalf("created host %q: read id=%q, want %q", host, got, id)
	}
	deleteHost(t, client, id, host)
	if got := hostID(t, client, host); got != "" {
		t.Fatalf("deleted host %q still listed with id %q", host, got)
	}
}

func TestConfigUI_CertificateUpload(t *testing.T) {
	client := integrationClient(t)
	const host = "certificate-test.localhost"
	certPEM, keyPEM := newHostCertificate(t, host)
	id := createHost(t, client, host, map[string][]byte{
		"serverCertFile": certPEM,
		"serverKeyFile":  keyPEM,
	})
	t.Cleanup(func() { deleteHost(t, client, id, host) })

	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certPEM) {
		t.Fatal("could not load generated test certificate")
	}
	// Wait for the live listener to switch to the configuration just acknowledged by the UI.
	// Trust only this test's newly generated certificate, and verify the virtual host via SNI.
	config := &tls.Config{RootCAs: roots, ServerName: host, MinVersion: tls.VersionTLS12}
	deadline := time.Now().Add(5 * time.Second)
	dialer := &net.Dialer{Timeout: time.Second}
	for {
		conn, err := tls.DialWithDialer(dialer, "tcp", "localhost:443", config)
		if err == nil {
			_ = conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("TLS listener did not serve the uploaded certificate for %q: %v", host, err)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func createHost(t *testing.T, client *http.Client, host string, files map[string][]byte) string {
	t.Helper()
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	fields := map[string]string{"from": host, "scheme": "http", "hostName": "backend-8080", "port": "80"}
	if len(files) != 0 {
		fields["useServerCert"] = "on"
	}
	for name, value := range fields {
		if err := form.WriteField(name, value); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range files {
		part, err := form.CreateFormFile(name, name+".pem")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := form.Close(); err != nil {
		t.Fatal(err)
	}
	res, err := client.Post(configUI+"/api/virtualhosts", form.FormDataContentType(), &body)
	if err != nil {
		t.Fatal(err)
	}
	var reply struct {
		Success bool
		ID      string
	}
	decodeOK(t, res, &reply)
	if !reply.Success || reply.ID == "" {
		t.Fatalf("creation response did not confirm a new virtual host: %+v", reply)
	}
	return reply.ID
}

func hostID(t *testing.T, client *http.Client, from string) string {
	t.Helper()
	res, err := client.Get(configUI + "/api/virtualhosts")
	if err != nil {
		t.Fatal(err)
	}
	var reply struct {
		Success      bool
		VirtualHosts []struct{ From, ID string }
	}
	decodeOK(t, res, &reply)
	if !reply.Success || len(reply.VirtualHosts) == 0 {
		t.Fatal("list response did not contain the integration virtual hosts")
	}
	for _, host := range reply.VirtualHosts {
		if host.From == from {
			return host.ID
		}
	}
	return ""
}

func deleteHost(t *testing.T, client *http.Client, id, host string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, configUI+"/api/virtualhosts/"+id, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var reply struct{ Success bool }
	decodeOK(t, res, &reply)
	if !reply.Success {
		t.Fatalf("delete response did not confirm id %s", id)
	}

	// The configuration API and TLS listener run independently. Observe retirement
	// using the fixture's localhost TLS endpoint and the removed HTTP host header.
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, DisableKeepAlives: true}
	defer transport.CloseIdleConnections()
	proxy := &http.Client{
		Transport: transport, Timeout: time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	probe, err := http.NewRequest(http.MethodGet, "https://localhost/", nil)
	if err != nil {
		t.Fatal(err)
	}
	probe.Host = host
	deadline := time.Now().Add(5 * time.Second)
	var status int
	for {
		res, err := proxy.Do(probe)
		if err == nil {
			status = res.StatusCode
			_ = res.Body.Close()
			if status == http.StatusNotFound {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("route %q did not retire: status=%d err=%v", host, status, err)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func decodeOK(t *testing.T, res *http.Response, dst any) {
	t.Helper()
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		t.Fatalf("%s %s: %s", res.Request.Method, res.Request.URL.Path, res.Status)
	}
	if ct := res.Header.Get("Content-Type"); strings.SplitN(ct, ";", 2)[0] != "application/json" {
		t.Fatalf("%s: expected JSON, got %q", res.Request.URL.Path, ct)
	}
	if err := json.NewDecoder(res.Body).Decode(dst); err != nil {
		t.Fatal(err)
	}
}
