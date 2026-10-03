package domain

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestWebVirtualHostRejectsSpoofedForwardingHeaders(t *testing.T) {
	// Arrange
	host, received := securityProxyFixture(t)
	request := spoofedProxyRequest()
	originalHeaders := request.Header.Clone()
	response := httptest.NewRecorder()
	// Act
	host.ServeHTTP(response, request)
	// Assert
	require.Equal(t, http.StatusNoContent, response.Code)
	backend := <-received
	assert.Equal(t, "192.0.2.9", backend.Header.Get("X-Forwarded-For"))
	assert.Equal(t, "public.example", backend.Header.Get("X-Forwarded-Host"))
	assert.Equal(t, "https", backend.Header.Get("X-Forwarded-Proto"))
	assert.Empty(t, backend.Header.Get("Forwarded"))
	assert.Empty(t, backend.Header.Get("X-Hop"))
	assert.Equal(t, "value", backend.Header.Get("X-Custom"))
	assert.Equal(t, originalHeaders, request.Header)
	assert.Equal(t, "/api/users", backend.URL.Path)
	assert.Equal(t, "good=1", backend.URL.RawQuery)
}

func securityProxyFixture(t *testing.T) (*WebVirtualHost, <-chan *http.Request) {
	t.Helper()
	received := make(chan *http.Request, 1)
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Clone(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(backend.Close)
	address, err := url.Parse(backend.URL)
	require.NoError(t, err)
	name, portText, err := net.SplitHostPort(address.Host)
	require.NoError(t, err)
	port, err := strconv.Atoi(portText)
	require.NoError(t, err)
	logger := &mocks.MockLogger{}
	logger.On("Info", mock.Anything).Return()
	logger.On("GetErrorLogger").Return(log.New(io.Discard, "", 0))
	host := &WebVirtualHost{}
	host.Scheme, host.HostName, host.Port = "http", name, uint(port)
	host.Path, host.pathToDelete, host.logger = "api", "/v1/", logger
	return host, received
}

func spoofedProxyRequest() *http.Request {
	request := httptest.NewRequest(http.MethodGet, "https://public.example/v1/users?good=1&bad=%zz", nil)
	request.RemoteAddr = "192.0.2.9:3210"
	request.Header.Set("Forwarded", "for=attacker")
	request.Header.Set("X-Forwarded-For", "203.0.113.66")
	request.Header.Set("X-Forwarded-Host", "attacker.example")
	request.Header.Set("X-Forwarded-Proto", "attacker")
	request.Header.Set("Connection", "X-Hop, X-Forwarded-Proto")
	request.Header.Set("X-Hop", "must-not-arrive")
	request.Header.Set("X-Custom", "value")
	return request
}
