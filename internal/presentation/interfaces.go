package presentation

import (
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	certificates "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/certificates"
)

// IVirtualHostService define la responsabilidad de gestionar operaciones de virtual hosts
type IVirtualHostService interface {
	CreateVirtualHost(r *http.Request, config *domain.Config) (interface{}, error)
	UpdateVirtualHost(r *http.Request, id string, config *domain.Config) (interface{}, []string, []string, error)
	DeleteVirtualHost(id string, config *domain.Config) (string, error)
	GetVirtualHost(id string, config *domain.Config) (interface{}, string, error)
	GetVirtualHosts(config *domain.Config) ([]domain.IVirtualHost, error)
	CleanupUnusedCertificates(config *domain.Config, oldServerPaths, oldClientPaths []string)
}

// ICertificateService define la responsabilidad de gestionar certificados
type ICertificateService interface {
	HandleServerCertificates(r *http.Request, certDir string) (*certificates.CertificateDefs, error)
	HandleClientCertificates(r *http.Request, certDir string) (*certificates.CertificateDefs, error)
	HandleCertificateUpdates(r *http.Request, certDir string, oldVH interface{}) (*certificates.CertificateDefs, *certificates.CertificateDefs, []string, []string, error)
	CollectCertPathsFromVH(config *domain.Config, id string) ([]string, []string)
	CollectAllCertsInUse(config *domain.Config) map[string]bool
	DeleteUnusedCertFiles(serverCertPaths, clientCertPaths []string, certsInUse map[string]bool)
	CleanupCertDirectory(config *domain.Config, deletedFrom string)
}

// IFileService define la responsabilidad de gestionar operaciones de archivo
type IFileService interface {
	CreateCertDirectory(config *domain.Config, from string) (string, error)
	SaveCertificateFromForm(r *http.Request, fieldName, destPath string) error
}

// IFormHandler define la responsabilidad de procesar formularios HTTP
type IFormHandler interface {
	ParseVirtualHostForm(r *http.Request, defaultVH *domain.WebVirtualHost) (string, uint, error)
}
