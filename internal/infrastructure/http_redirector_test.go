package infrastructure

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/janmbaco/go-reverseproxy-ssl/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of Logger
type MockLogger = mocks.MockLogger

// MockVirtualHost is a mock implementation of IVirtualHost
type MockVirtualHost struct {
	mock.Mock
	from string
}

func (m *MockVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.Called(rw, req)
}

func (m *MockVirtualHost) GetID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockVirtualHost) GetFrom() string {
	return m.from
}

func (m *MockVirtualHost) GetHostToReplace() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockVirtualHost) GetURLToReplace() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockVirtualHost) GetURL() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockVirtualHost) GetAuthorizedCAs() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockVirtualHost) GetServerCertificate() domain.CertificateProvider {
	args := m.Called()
	return args.Get(0).(domain.CertificateProvider)
}

func (m *MockVirtualHost) GetHostName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockVirtualHost) SetURLToReplace() {
	m.Called()
}

func (m *MockVirtualHost) EnsureID() {
	m.Called()
}

func TestNewHTTPRedirector_WhenCalled_ThenReturnsRedirector(t *testing.T) {
	// Arrange
	mockLogger := &MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()

	// Act
	redirector := NewHTTPRedirector(mockLogger)

	// Assert
	assert.NotNil(t, redirector)
	assert.NotNil(t, redirector.mux)
	assert.NotNil(t, redirector.redirectRules)
	assert.Equal(t, mockLogger, redirector.logger)
	assert.Equal(t, ":80", redirector.server.Addr)
}

func TestHTTPRedirector_UpdateRedirectRules_WhenCalled_ThenUpdatesRules(t *testing.T) {
	// Arrange
	mockLogger := &MockLogger{}
	mockLogger.On("Info", mock.Anything).Return() // Allow Info calls
	redirector := NewHTTPRedirector(mockLogger)
	vhosts := []domain.IVirtualHost{
		&MockVirtualHost{from: "example.com"},
		&MockVirtualHost{from: "test.com"},
	}

	// Act
	redirector.UpdateRedirectRules(vhosts)

	// Assert
	assert.Len(t, redirector.redirectRules, 2)
	assert.Equal(t, "https://example.com", redirector.redirectRules["example.com"])
	assert.Equal(t, "https://test.com", redirector.redirectRules["test.com"])
	mockLogger.AssertExpectations(t)
}

func TestHTTPRedirector_handleRedirect_WhenRuleExists_ThenRedirectsToRule(t *testing.T) {
	// Arrange
	mockLogger := &MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	redirector := NewHTTPRedirector(mockLogger)
	redirector.redirectRules["example.com"] = "https://secure.example.com"

	req := httptest.NewRequest("GET", "/path", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()

	// Act
	redirector.handleRedirect(w, req)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://secure.example.com", w.Header().Get("Location"))
}

func TestHTTPRedirector_handleRedirect_WhenNoRule_ThenGenericRedirect(t *testing.T) {
	// Arrange
	mockLogger := &MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()
	redirector := NewHTTPRedirector(mockLogger)

	req := httptest.NewRequest("GET", "/path?query=1", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()

	// Act
	redirector.handleRedirect(w, req)

	// Assert
	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://example.com/path?query=1", w.Header().Get("Location"))
}
