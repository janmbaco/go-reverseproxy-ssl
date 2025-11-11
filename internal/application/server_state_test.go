package application

import (
	"net/http"
	"sync"
	"testing"

	"github.com/janmbaco/go-reverseproxy-ssl/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestServerState_UpdateMux_WhenCalled_ThenUpdatesMux(t *testing.T) {
	// Arrange
	state := &ServerState{}
	mux := http.NewServeMux()

	// Act
	state.UpdateMux(mux)

	// Assert
	assert.Equal(t, mux, state.GetMux())
}

func TestServerState_UpdateCertMgr_WhenCalled_ThenUpdatesCertMgr(t *testing.T) {
	// Arrange
	state := &ServerState{}
	var certMgr domain.CertificateManager

	// Act
	state.UpdateCertMgr(certMgr)

	// Assert
	assert.Equal(t, certMgr, state.GetCertMgr())
}

func TestServerState_UpdateConfig_WhenCalled_ThenUpdatesConfigWithCopy(t *testing.T) {
	// Arrange
	state := &ServerState{}
	config := &domain.Config{
		DefaultHost: "example.com",
	}

	// Act
	state.UpdateConfig(config)

	// Assert
	retrieved := state.GetConfig()
	assert.NotNil(t, retrieved)
	assert.Equal(t, "example.com", retrieved.DefaultHost)
	assert.NotSame(t, config, retrieved) // Should be a copy
}

func TestServerState_GetConfig_WhenConfigNil_ThenReturnsNil(t *testing.T) {
	// Arrange
	state := &ServerState{}

	// Act
	result := state.GetConfig()

	// Assert
	assert.Nil(t, result)
}

func TestServerState_Concurrency_WhenMultipleGoroutines_ThenThreadSafe(t *testing.T) {
	// Arrange
	state := &ServerState{}
	var wg sync.WaitGroup

	// Act
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mux := http.NewServeMux()
			state.UpdateMux(mux)
			_ = state.GetMux()
		}()
	}
	wg.Wait()

	// Assert
	// No race conditions, test passes if no panic
	assert.NotNil(t, state)
}
