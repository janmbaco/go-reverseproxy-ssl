package domain

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/janmbaco/go-reverseproxy-ssl/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestVirtualHostBase_EnsureID_WhenIDIsEmpty_ThenGeneratesUUID(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		From: "example.com",
	}

	// Act
	vh.EnsureID()

	// Assert
	assert.NotEmpty(t, vh.ID)
	assert.Len(t, vh.ID, 36) // UUID length
}

func TestVirtualHostBase_EnsureID_WhenIDAlreadySet_ThenDoesNotChange(t *testing.T) {
	// Arrange
	existingID := "existing-id"
	vh := &VirtualHostBase{
		ID:   existingID,
		From: "example.com",
	}

	// Act
	vh.EnsureID()

	// Assert
	assert.Equal(t, existingID, vh.ID)
}

func TestVirtualHostBase_GetID_WhenCalled_ThenReturnsID(t *testing.T) {
	// Arrange
	expectedID := "test-id"
	vh := &VirtualHostBase{
		ID:   expectedID,
		From: "example.com",
	}

	// Act
	actualID := vh.GetID()

	// Assert
	assert.Equal(t, expectedID, actualID)
}

func TestVirtualHostBase_GetFrom_WhenCalled_ThenReturnsFrom(t *testing.T) {
	// Arrange
	expectedFrom := "example.com"
	vh := &VirtualHostBase{
		From: expectedFrom,
	}

	// Act
	actualFrom := vh.GetFrom()

	// Assert
	assert.Equal(t, expectedFrom, actualFrom)
}

func TestVirtualHostBase_SetURLToReplace_WhenFromHasPath_ThenSetsCorrectValues(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		From: "example.com/path/",
	}

	// Act
	vh.SetURLToReplace()

	// Assert
	assert.Equal(t, "example.com/path/", vh.urlToReplace)
	assert.Equal(t, "example.com", vh.hostToReplace)
	assert.Equal(t, "/path/", vh.pathToDelete)
}

func TestVirtualHostBase_SetURLToReplace_WhenFromNoPath_ThenSetsCorrectValues(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		From: "example.com",
	}

	// Act
	vh.SetURLToReplace()

	// Assert
	assert.Equal(t, "example.com/", vh.urlToReplace)
	assert.Equal(t, "example.com", vh.hostToReplace)
	assert.Equal(t, "", vh.pathToDelete)
}

func TestVirtualHostBase_GetHostToReplace_WhenCalled_ThenReturnsHostToReplace(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		hostToReplace: "example.com",
	}

	// Act
	result := vh.GetHostToReplace()

	// Assert
	assert.Equal(t, "example.com", result)
}

func TestVirtualHostBase_GetURLToReplace_WhenCalled_ThenReturnsURLToReplace(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		urlToReplace: "example.com/",
	}

	// Act
	result := vh.GetURLToReplace()

	// Assert
	assert.Equal(t, "example.com/", result)
}

func TestVirtualHostBase_GetURL_WhenCalled_ThenReturnsFormattedURL(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Scheme:   "https",
		HostName: "example.com",
		Port:     443,
		Path:     "api",
	}

	// Act
	result := vh.GetURL()

	// Assert
	assert.Equal(t, "'https://example.com:443/api'", result)
}

func TestVirtualHostBase_GetAuthorizedCAs_WhenServerCertificateNil_ThenReturnsEmpty(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{}

	// Act
	result := vh.GetAuthorizedCAs()

	// Assert
	assert.Empty(t, result)
}

func TestVirtualHostBase_GetHostName_WhenCalled_ThenReturnsHostNameWithPort(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		HostName: "example.com",
		Port:     8080,
	}

	// Act
	result := vh.GetHostName()

	// Assert
	assert.Equal(t, "example.com:8080", result)
}

func TestVirtualHostBase_getPath_WhenPathEmptyAndNoPathToDelete_ThenReturnsOriginalPath(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Path:         "",
		pathToDelete: "",
	}

	// Act
	result := vh.getPath("/api/users")

	// Assert
	assert.Equal(t, "/api/users", result)
}

func TestVirtualHostBase_getPath_WhenPathSetAndNoPathToDelete_ThenAddsPathPrefix(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Path:         "api",
		pathToDelete: "",
	}

	// Act
	result := vh.getPath("/users")

	// Assert
	assert.Equal(t, "/api/users", result)
}

func TestVirtualHostBase_getPath_WhenPathSetAndPathToDelete_ThenReplacesAndAddsPrefix(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Path:         "v1",
		pathToDelete: "/oldapi/",
	}

	// Act
	result := vh.getPath("/oldapi/users/123")

	// Assert
	assert.Equal(t, "/v1/users/123", result)
}

func TestVirtualHostBase_getPath_WhenPathWithTrailingSlash_ThenHandlesCorrectly(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Path:         "api/",
		pathToDelete: "/v1/",
	}

	// Act
	result := vh.getPath("/v1/users")

	// Assert
	assert.Equal(t, "/api/users", result)
}

func TestVirtualHostBase_getPath_WhenPathToDeleteAtEnd_ThenReplacesCorrectly(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Path:         "new",
		pathToDelete: "/old/",
	}

	// Act
	result := vh.getPath("/old/")

	// Assert
	assert.Equal(t, "/new/", result)
}

func TestVirtualHostBase_getPath_WhenDoubleSlashes_ThenNormalizes(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Path:         "",
		pathToDelete: "/api/",
	}

	// Act
	result := vh.getPath("/api//users//123")

	// Assert
	assert.Equal(t, "/users/123", result)
}

func TestVirtualHostBase_getPath_WhenComplexPathReplacement_ThenWorksCorrectly(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		Path:         "microservice",
		pathToDelete: "/legacy/service/",
	}

	// Act
	result := vh.getPath("/legacy/service/users/profile")

	// Assert
	assert.Equal(t, "/microservice/users/profile", result)
}

func TestVirtualHostBase_SetURLToReplace_WhenFromWithMultiplePaths_ThenSetsCorrectPathToDelete(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		From: "api.example.com/v1/users",
	}

	// Act
	vh.SetURLToReplace()

	// Assert
	assert.Equal(t, "api.example.com/v1/users/", vh.urlToReplace)
	assert.Equal(t, "api.example.com", vh.hostToReplace)
	assert.Equal(t, "/v1/users/", vh.pathToDelete)
}

func TestVirtualHostBase_SetURLToReplace_WhenFromEndsWithSlash_ThenHandlesCorrectly(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		From: "example.com/api/",
	}

	// Act
	vh.SetURLToReplace()

	// Assert
	assert.Equal(t, "example.com/api/", vh.urlToReplace)
	assert.Equal(t, "example.com", vh.hostToReplace)
	assert.Equal(t, "/api/", vh.pathToDelete)
}

func TestVirtualHostBase_SetURLToReplace_WhenFromNoTrailingSlash_ThenAddsSlashToPathToDelete(t *testing.T) {
	// Arrange
	vh := &VirtualHostBase{
		From: "example.com/api",
	}

	// Act
	vh.SetURLToReplace()

	// Assert
	assert.Equal(t, "example.com/api/", vh.urlToReplace)
	assert.Equal(t, "example.com", vh.hostToReplace)
	assert.Equal(t, "/api/", vh.pathToDelete)
}

func TestVirtualHostBase_RoutingScenario_ApiGateway_ThenRoutesCorrectly(t *testing.T) {
	// Arrange: API Gateway scenario
	// From: "api.example.com/v1" -> HostName: "internal-api", Path: "api"
	vh := &VirtualHostBase{
		From:     "api.example.com/v1",
		Scheme:   "http",
		HostName: "internal-api",
		Port:     8080,
		Path:     "api",
	}
	vh.SetURLToReplace()

	// Act: Request comes to "api.example.com/v1/users/123"
	result := vh.getPath("/v1/users/123")

	// Assert: Should route to "internal-api:8080/api/users/123"
	assert.Equal(t, "/api/users/123", result)
	expectedHost := "internal-api:8080"
	assert.Equal(t, expectedHost, vh.GetHostName())
}

func TestVirtualHostBase_RoutingScenario_Microservice_ThenRoutesCorrectly(t *testing.T) {
	// Arrange: Microservice scenario
	// From: "example.com/users" -> HostName: "user-service", Path: ""
	vh := &VirtualHostBase{
		From:     "example.com/users",
		Scheme:   "http",
		HostName: "user-service",
		Port:     3000,
		Path:     "",
	}
	vh.SetURLToReplace()

	// Act: Request comes to "example.com/users/profile/456"
	result := vh.getPath("/users/profile/456")

	// Assert: Should route to "user-service:3000/profile/456"
	assert.Equal(t, "/profile/456", result)
	expectedHost := "user-service:3000"
	assert.Equal(t, expectedHost, vh.GetHostName())
}

func TestVirtualHostBase_RoutingScenario_VersionedApi_ThenRoutesCorrectly(t *testing.T) {
	// Arrange: Versioned API scenario
	// From: "api.company.com" -> HostName: "api-backend", Path: "v2"
	vh := &VirtualHostBase{
		From:     "api.company.com",
		Scheme:   "https",
		HostName: "api-backend",
		Port:     443,
		Path:     "v2",
	}
	vh.SetURLToReplace()

	// Act: Request comes to "api.company.com/orders"
	result := vh.getPath("/orders")

	// Assert: Should route to "api-backend:443/v2/orders"
	assert.Equal(t, "/v2/orders", result)
	expectedHost := "api-backend:443"
	assert.Equal(t, expectedHost, vh.GetHostName())
}

func TestVirtualHostBase_RoutingScenario_LegacyMigration_ThenRoutesCorrectly(t *testing.T) {
	// Arrange: Legacy migration scenario
	// From: "old-api.com/legacy" -> HostName: "new-api", Path: "modern"
	vh := &VirtualHostBase{
		From:     "old-api.com/legacy",
		Scheme:   "https",
		HostName: "new-api",
		Port:     9000,
		Path:     "modern",
	}
	vh.SetURLToReplace()

	// Act: Request comes to "old-api.com/legacy/endpoint"
	result := vh.getPath("/legacy/endpoint")

	// Assert: Should route to "new-api:9000/modern/endpoint"
	assert.Equal(t, "/modern/endpoint", result)
	expectedHost := "new-api:9000"
	assert.Equal(t, expectedHost, vh.GetHostName())
}

func TestVirtualHostBase_redirectRequest_WhenCalled_ThenTransformsRequestCorrectly(t *testing.T) {
	// Arrange
	mockLogger := &mocks.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()

	vh := &VirtualHostBase{
		Scheme:       "https",
		HostName:     "backend.com",
		Port:         8443,
		Path:         "api",
		pathToDelete: "/v1/",
		logger:       mockLogger,
	}

	inReq := &http.Request{
		Method: "GET",
		URL:    &url.URL{Path: "/v1/users/123", RawQuery: "filter=active"},
		Header: http.Header{"X-Custom": []string{"value"}},
		Host:   "frontend.com",
	}

	outReq := &http.Request{
		Method: "GET",
		URL:    &url.URL{},
		Header: http.Header{},
	}

	// Act
	vh.redirectRequest(outReq, inReq, true)

	// Assert
	assert.Equal(t, "https", outReq.URL.Scheme)
	assert.Equal(t, "backend.com:8443", outReq.URL.Host)
	assert.Equal(t, "/api/users/123", outReq.URL.Path)
	assert.Equal(t, "filter=active", outReq.URL.RawQuery)
	assert.Equal(t, "value", outReq.Header.Get("X-Custom"))
	assert.Equal(t, "https", outReq.Header.Get("X-Forwarded-Proto"))
	mockLogger.AssertExpectations(t)
}

func TestVirtualHostBase_redirectRequest_WhenNoXForwardedHeader_ThenDoesNotSetHeader(t *testing.T) {
	// Arrange
	mockLogger := &mocks.MockLogger{}
	mockLogger.On("Info", mock.Anything).Return()

	vh := &VirtualHostBase{
		Scheme:       "http",
		HostName:     "backend.com",
		Port:         8080,
		Path:         "",
		pathToDelete: "",
		logger:       mockLogger,
	}

	inReq := &http.Request{
		Method: "POST",
		URL:    &url.URL{Path: "/data", RawQuery: ""},
		Header: http.Header{"Content-Type": []string{"application/json"}},
		Host:   "frontend.com",
	}

	outReq := &http.Request{
		Method: "POST",
		URL:    &url.URL{},
		Header: http.Header{},
	}

	// Act
	vh.redirectRequest(outReq, inReq, false)

	// Assert
	assert.Equal(t, "http", outReq.URL.Scheme)
	assert.Equal(t, "backend.com:8080", outReq.URL.Host)
	assert.Equal(t, "/data", outReq.URL.Path)
	assert.Equal(t, "", outReq.URL.RawQuery)
	assert.Equal(t, "application/json", outReq.Header.Get("Content-Type"))
	assert.Empty(t, outReq.Header.Get("X-Forwarded-Proto"))
	mockLogger.AssertExpectations(t)
}

func TestVirtualHostBase_EndToEndRouting_ComplexScenario_ThenRoutesCorrectly(t *testing.T) {
	// Arrange: Complex routing scenario
	// Frontend: api.example.com/v2/service -> Backend: internal-service.company.com:9000/api/v2/service
	vh := &VirtualHostBase{
		From:     "api.example.com/v2/service",
		Scheme:   "https",
		HostName: "internal-service.company.com",
		Port:     9000,
		Path:     "api",
	}
	vh.SetURLToReplace()

	// Verify parsing
	assert.Equal(t, "api.example.com", vh.hostToReplace)
	assert.Equal(t, "/v2/service/", vh.pathToDelete)

	// Act: Simulate incoming request to "api.example.com/v2/service/users/list"
	incomingPath := "/v2/service/users/list"
	transformedPath := vh.getPath(incomingPath)

	// Assert: Should route to "internal-service.company.com:9000/api/users/list"
	assert.Equal(t, "/api/users/list", transformedPath)
	assert.Equal(t, "internal-service.company.com:9000", vh.GetHostName())
}

func TestVirtualHostBase_PathNormalization_EdgeCases_ThenHandlesCorrectly(t *testing.T) {
	testCases := []struct {
		name         string
		vhPath       string
		pathToDelete string
		inputPath    string
		expected     string
	}{
		{
			name:         "Empty path with double slashes",
			vhPath:       "",
			pathToDelete: "",
			inputPath:    "//api//users//",
			expected:     "/api/users/",
		},
		{
			name:         "Path replacement with double slashes",
			vhPath:       "v1",
			pathToDelete: "/old/",
			inputPath:    "/old//users//123",
			expected:     "/v1/users/123",
		},
		{
			name:         "Root path replacement",
			vhPath:       "api",
			pathToDelete: "/",
			inputPath:    "/users",
			expected:     "/api/users",
		},
		{
			name:         "Complex nested paths",
			vhPath:       "service/v2",
			pathToDelete: "/api/v1/",
			inputPath:    "/api/v1/users/profile/settings",
			expected:     "/service/v2/users/profile/settings",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			vh := &VirtualHostBase{
				Path:         tc.vhPath,
				pathToDelete: tc.pathToDelete,
			}

			// Act
			result := vh.getPath(tc.inputPath)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}
