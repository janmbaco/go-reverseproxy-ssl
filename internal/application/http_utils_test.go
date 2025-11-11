package application

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedirectToWWW_WhenHostnameHasWWW_ThenRegistersRedirect(t *testing.T) {
	// Arrange
	mux := http.NewServeMux()
	hostname := "www.example.com"

	// Act
	RedirectToWWW(hostname, mux)

	// Assert
	req := httptest.NewRequest("GET", "/path", nil)
	req.Host = "example.com"
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "https://www.example.com/path", w.Header().Get("Location"))
}

func TestRedirectToWWW_WhenHostnameNoWWW_ThenDoesNothing(t *testing.T) {
	// Arrange
	mux := http.NewServeMux()
	hostname := "example.com"

	// Act
	RedirectToWWW(hostname, mux)

	// Assert
	// Should not have registered anything, so 404
	req := httptest.NewRequest("GET", "http://example.com/path", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
