package application

import (
	"net/http"
	"sync"

	"github.com/janmbaco/go-reverseproxy-ssl/internal/domain"
	"github.com/jinzhu/copier"
)

type ServerState struct {
	mux     *http.ServeMux
	certMgr domain.CertificateManager
	config  *domain.Config
	mutex   sync.RWMutex
}

func (s *ServerState) UpdateMux(mux *http.ServeMux) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.mux = mux
}

func (s *ServerState) GetMux() *http.ServeMux {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.mux
}

func (s *ServerState) UpdateCertMgr(certMgr domain.CertificateManager) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.certMgr = certMgr
}

func (s *ServerState) GetCertMgr() domain.CertificateManager {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.certMgr
}

func (s *ServerState) UpdateConfig(config *domain.Config) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a deep copy to ensure the internal state is not affected by external modifications
	if config == nil {
		s.config = nil
		return
	}

	configCopy := &domain.Config{}
	err := copier.Copy(configCopy, config)
	if err != nil {
		// If copying fails, we don't update the config
		// In production, this should be logged
		return
	}

	s.config = configCopy
}

func (s *ServerState) GetConfig() *domain.Config {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Return a deep copy to prevent external modifications
	if s.config == nil {
		return nil
	}

	// Create a deep copy using copier library
	configCopy := &domain.Config{}
	err := copier.Copy(configCopy, s.config)
	if err != nil {
		// If copying fails, return nil to prevent returning shared state
		// In production, this should be logged
		return nil
	}

	return configCopy
}
