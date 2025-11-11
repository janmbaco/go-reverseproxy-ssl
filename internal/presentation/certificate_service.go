package presentation

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	certificates "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/certificates"
)

// CertificateService implementa la responsabilidad de gestionar certificados
type CertificateService struct {
	fileService IFileService
	logger      domain.Logger
}

// NewCertificateService crea una nueva instancia de CertificateService
func NewCertificateService(fileService IFileService, logger domain.Logger) ICertificateService {
	return &CertificateService{
		fileService: fileService,
		logger:      logger,
	}
}

// HandleServerCertificates implementa ICertificateService.HandleServerCertificates
func (cs *CertificateService) HandleServerCertificates(r *http.Request, certDir string) (*certificates.CertificateDefs, error) {
	var serverCert *certificates.CertificateDefs

	useServerCert := r.FormValue("useServerCert") == "on"
	if !useServerCert {
		return nil, nil
	}

	// Save server certificate if provided
	if _, header, err := r.FormFile("serverCertFile"); err == nil {
		serverCertPath := filepath.Join(certDir, header.Filename)
		if err := cs.fileService.SaveCertificateFromForm(r, "serverCertFile", serverCertPath); err != nil {
			cs.logger.Error(fmt.Sprintf("Failed to save server cert: %v", err))
			return nil, err
		}
		cs.logger.Info(fmt.Sprintf("Saved server certificate to %s", serverCertPath))

		if serverCert == nil {
			serverCert = &certificates.CertificateDefs{}
		}
		serverCert.PublicKey = serverCertPath
	}

	// Save server key if provided
	if _, header, err := r.FormFile("serverKeyFile"); err == nil {
		serverKeyPath := filepath.Join(certDir, header.Filename)
		if err := cs.fileService.SaveCertificateFromForm(r, "serverKeyFile", serverKeyPath); err != nil {
			cs.logger.Error(fmt.Sprintf("Failed to save server key: %v", err))
			return nil, err
		}
		cs.logger.Info(fmt.Sprintf("Saved server key to %s", serverKeyPath))

		if serverCert == nil {
			serverCert = &certificates.CertificateDefs{}
		}
		serverCert.PrivateKey = serverKeyPath
	}

	return serverCert, nil
}

// HandleClientCertificates implementa ICertificateService.HandleClientCertificates
func (cs *CertificateService) HandleClientCertificates(r *http.Request, certDir string) (*certificates.CertificateDefs, error) {
	var clientCert *certificates.CertificateDefs

	useClientCert := r.FormValue("useClientCert") == "on"
	if !useClientCert {
		return nil, nil
	}

	// Save client certificate if provided
	if _, header, err := r.FormFile("clientCertFile"); err == nil {
		clientCertPath := filepath.Join(certDir, header.Filename)
		if err := cs.fileService.SaveCertificateFromForm(r, "clientCertFile", clientCertPath); err != nil {
			cs.logger.Error(fmt.Sprintf("Failed to save client cert: %v", err))
			return nil, err
		}
		cs.logger.Info(fmt.Sprintf("Saved client CA certificate to %s", clientCertPath))

		if clientCert == nil {
			clientCert = &certificates.CertificateDefs{}
		}
		clientCert.CaPem = []string{clientCertPath}
	}

	return clientCert, nil
}

// HandleCertificateUpdates implementa ICertificateService.HandleCertificateUpdates
func (cs *CertificateService) HandleCertificateUpdates(r *http.Request, certDir string, oldVH interface{}) (*certificates.CertificateDefs, *certificates.CertificateDefs, []string, []string, error) {
	// Collect old certificate paths for cleanup
	oldServerPaths, oldClientPaths := cs.collectOldCertificatePaths(oldVH)

	// Initialize certificate copies from existing ones
	serverCert, clientCert := cs.initializeCertificateCopies(oldVH)

	// Handle certificate updates
	if err := cs.handleServerCertificateUpdate(r, certDir, &serverCert); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := cs.handleServerKeyUpdate(r, certDir, &serverCert); err != nil {
		return nil, nil, nil, nil, err
	}

	if err := cs.handleClientCertificateUpdate(r, certDir, &clientCert); err != nil {
		return nil, nil, nil, nil, err
	}

	return serverCert, clientCert, oldServerPaths, oldClientPaths, nil
}

// CollectCertPathsFromVH implementa ICertificateService.CollectCertPathsFromVH
func (cs *CertificateService) CollectCertPathsFromVH(config *domain.Config, id string) ([]string, []string) {
	// Check WebVirtualHosts first
	if serverPaths, clientPaths := cs.collectCertPathsFromWebHosts(config.WebVirtualHosts, id); len(serverPaths) > 0 || len(clientPaths) > 0 {
		return serverPaths, clientPaths
	}

	// Check GrpcWebVirtualHosts
	return cs.collectCertPathsFromGrpcHosts(config.GrpcWebVirtualHosts, id)
}

// collectCertPathsFromWebHosts collects certificate paths from WebVirtualHosts
func (cs *CertificateService) collectCertPathsFromWebHosts(hosts []*domain.WebVirtualHost, id string) ([]string, []string) {
	for _, vh := range hosts {
		if vh.GetID() == id {
			return cs.extractCertPaths(vh.ServerCertificate, vh.ClientCertificate)
		}
	}
	return nil, nil
}

// collectCertPathsFromGrpcHosts collects certificate paths from GrpcWebVirtualHosts
func (cs *CertificateService) collectCertPathsFromGrpcHosts(hosts []*domain.GrpcWebVirtualHost, id string) ([]string, []string) {
	for _, vh := range hosts {
		if vh.GetID() == id {
			return cs.extractCertPaths(vh.ServerCertificate, vh.ClientCertificate)
		}
	}
	return nil, nil
}

// extractCertPaths extracts certificate paths from server and client certificates
func (cs *CertificateService) extractCertPaths(serverCert *certificates.CertificateDefs, clientCert *certificates.CertificateDefs) ([]string, []string) {
	var serverPaths []string
	var clientPaths []string

	if serverCert != nil {
		if serverCert.PublicKey != "" {
			serverPaths = append(serverPaths, serverCert.PublicKey)
		}
		if serverCert.PrivateKey != "" {
			serverPaths = append(serverPaths, serverCert.PrivateKey)
		}
	}

	if clientCert != nil && len(clientCert.CaPem) > 0 {
		clientPaths = append(clientPaths, clientCert.CaPem...)
	}

	return serverPaths, clientPaths
}

// CollectAllCertsInUse implementa ICertificateService.CollectAllCertsInUse
func (cs *CertificateService) CollectAllCertsInUse(config *domain.Config) map[string]bool {
	certsInUse := make(map[string]bool)

	// Check WebVirtualHosts
	for _, vh := range config.WebVirtualHosts {
		if vh.ServerCertificate != nil {
			if vh.ServerCertificate.PublicKey != "" {
				certsInUse[vh.ServerCertificate.PublicKey] = true
			}
			if vh.ServerCertificate.PrivateKey != "" {
				certsInUse[vh.ServerCertificate.PrivateKey] = true
			}
		}
		if vh.ClientCertificate != nil {
			for _, ca := range vh.ClientCertificate.CaPem {
				certsInUse[ca] = true
			}
		}
	}

	// Check GrpcWebVirtualHosts
	for _, vh := range config.GrpcWebVirtualHosts {
		if vh.ServerCertificate != nil {
			if vh.ServerCertificate.PublicKey != "" {
				certsInUse[vh.ServerCertificate.PublicKey] = true
			}
			if vh.ServerCertificate.PrivateKey != "" {
				certsInUse[vh.ServerCertificate.PrivateKey] = true
			}
		}
		if vh.ClientCertificate != nil {
			for _, ca := range vh.ClientCertificate.CaPem {
				certsInUse[ca] = true
			}
		}
	}

	return certsInUse
}

// DeleteUnusedCertFiles implementa ICertificateService.DeleteUnusedCertFiles
func (cs *CertificateService) DeleteUnusedCertFiles(serverCertPaths, clientCertPaths []string, certsInUse map[string]bool) {
	for _, certPath := range serverCertPaths {
		if !certsInUse[certPath] {
			cs.logger.Info(fmt.Sprintf("Deleting unused certificate: %s", certPath))
			if err := os.Remove(certPath); err != nil {
				cs.logger.Error(fmt.Sprintf("Failed to delete certificate %s: %v", certPath, err))
			}
		} else {
			cs.logger.Info(fmt.Sprintf("Certificate %s is still in use, skipping deletion", certPath))
		}
	}

	for _, certPath := range clientCertPaths {
		if !certsInUse[certPath] {
			cs.logger.Info(fmt.Sprintf("Deleting unused client certificate: %s", certPath))
			if err := os.Remove(certPath); err != nil {
				cs.logger.Error(fmt.Sprintf("Failed to delete client certificate %s: %v", certPath, err))
			}
		} else {
			cs.logger.Info(fmt.Sprintf("Client certificate %s is still in use, skipping deletion", certPath))
		}
	}
}

// CleanupCertDirectory implementa ICertificateService.CleanupCertDirectory
func (cs *CertificateService) CleanupCertDirectory(config *domain.Config, deletedFrom string) {
	certBaseDir := config.CertDir
	if certBaseDir == "" {
		certBaseDir = "/app/certs"
	}
	safeName := cs.sanitizePathName(deletedFrom)
	certDir := filepath.Join(certBaseDir, safeName)

	cs.logger.Info(fmt.Sprintf("Checking certificate directory: %s", certDir))
	if entries, err := os.ReadDir(certDir); err == nil {
		if len(entries) == 0 {
			cs.logger.Info(fmt.Sprintf("Deleting empty certificate directory: %s", certDir))
			if err := os.Remove(certDir); err != nil {
				cs.logger.Error(fmt.Sprintf("Failed to delete directory %s: %v", certDir, err))
			}
		} else {
			cs.logger.Info(fmt.Sprintf("Directory %s still contains files, keeping it", certDir))
		}
	}
}

// sanitizePathName sanitizes a path name to prevent directory traversal attacks
func (cs *CertificateService) sanitizePathName(name string) string {
	// If name contains a URL-like format, extract just the host/domain part
	if strings.Contains(name, "://") {
		// Parse as URL to extract host
		if parts := strings.Split(name, "://"); len(parts) > 1 {
			hostAndPath := parts[1]
			// Take everything before the first '/' (path separator)
			if hostEnd := strings.Index(hostAndPath, "/"); hostEnd > 0 {
				name = hostAndPath[:hostEnd]
			} else {
				name = hostAndPath
			}
		}
	} else if strings.Contains(name, "/") {
		// If no protocol but has path, take only the domain part
		if hostEnd := strings.Index(name, "/"); hostEnd > 0 {
			name = name[:hostEnd]
		}
	}

	// Replace dangerous characters
	safeName := strings.ReplaceAll(name, "/", "-")
	safeName = strings.ReplaceAll(safeName, ":", "-")
	safeName = strings.ReplaceAll(safeName, "\\", "-")
	safeName = strings.ReplaceAll(safeName, "..", "-")
	safeName = strings.ReplaceAll(safeName, "<", "-")
	safeName = strings.ReplaceAll(safeName, ">", "-")
	safeName = strings.ReplaceAll(safeName, "|", "-")
	safeName = strings.ReplaceAll(safeName, "?", "-")
	safeName = strings.ReplaceAll(safeName, "*", "-")
	safeName = strings.ReplaceAll(safeName, "\"", "-")
	safeName = strings.ReplaceAll(safeName, "'", "-")

	// Remove any remaining dangerous sequences
	safeName = strings.ReplaceAll(safeName, "..", "")

	// Ensure it's not empty and doesn't start/end with dangerous chars
	safeName = strings.Trim(safeName, ".- ")

	// If empty after sanitization, use a default
	if safeName == "" {
		safeName = "default"
	}

	return safeName
}

// collectOldCertificatePaths collects certificate paths from any virtual host type
func (cs *CertificateService) collectOldCertificatePaths(oldVH interface{}) ([]string, []string) {
	var oldServerPaths []string
	var oldClientPaths []string

	// Try to extract certificates from the virtual host
	if vh, ok := oldVH.(*domain.WebVirtualHost); ok {
		if vh.ServerCertificate != nil {
			if vh.ServerCertificate.PublicKey != "" {
				oldServerPaths = append(oldServerPaths, vh.ServerCertificate.PublicKey)
			}
			if vh.ServerCertificate.PrivateKey != "" {
				oldServerPaths = append(oldServerPaths, vh.ServerCertificate.PrivateKey)
			}
		}
		if vh.ClientCertificate != nil && len(vh.ClientCertificate.CaPem) > 0 {
			oldClientPaths = append(oldClientPaths, vh.ClientCertificate.CaPem...)
		}
	} else if vh, ok := oldVH.(*domain.GrpcWebVirtualHost); ok {
		if vh.ServerCertificate != nil {
			if vh.ServerCertificate.PublicKey != "" {
				oldServerPaths = append(oldServerPaths, vh.ServerCertificate.PublicKey)
			}
			if vh.ServerCertificate.PrivateKey != "" {
				oldServerPaths = append(oldServerPaths, vh.ServerCertificate.PrivateKey)
			}
		}
		if vh.ClientCertificate != nil && len(vh.ClientCertificate.CaPem) > 0 {
			oldClientPaths = append(oldClientPaths, vh.ClientCertificate.CaPem...)
		}
	}

	return oldServerPaths, oldClientPaths
}

// initializeCertificateCopies initializes certificate copies from any virtual host type
func (cs *CertificateService) initializeCertificateCopies(oldVH interface{}) (*certificates.CertificateDefs, *certificates.CertificateDefs) {
	var serverCert *certificates.CertificateDefs
	var clientCert *certificates.CertificateDefs

	// Try to extract certificates from the virtual host
	if vh, ok := oldVH.(*domain.WebVirtualHost); ok {
		// Copy existing certificates
		if vh.ServerCertificate != nil {
			serverCert = &certificates.CertificateDefs{
				PublicKey:  vh.ServerCertificate.PublicKey,
				PrivateKey: vh.ServerCertificate.PrivateKey,
				CaPem:      vh.ServerCertificate.CaPem,
			}
		}
		if vh.ClientCertificate != nil {
			clientCert = &certificates.CertificateDefs{
				CaPem: vh.ClientCertificate.CaPem,
			}
		}
	} else if vh, ok := oldVH.(*domain.GrpcWebVirtualHost); ok {
		// Copy existing certificates
		if vh.ServerCertificate != nil {
			serverCert = &certificates.CertificateDefs{
				PublicKey:  vh.ServerCertificate.PublicKey,
				PrivateKey: vh.ServerCertificate.PrivateKey,
				CaPem:      vh.ServerCertificate.CaPem,
			}
		}
		if vh.ClientCertificate != nil {
			clientCert = &certificates.CertificateDefs{
				CaPem: vh.ClientCertificate.CaPem,
			}
		}
	}

	return serverCert, clientCert
}

func (cs *CertificateService) handleServerCertificateUpdate(r *http.Request, certDir string, serverCert **certificates.CertificateDefs) error {
	if _, header, err := r.FormFile("serverCertFile"); err == nil {
		serverCertPath := filepath.Join(certDir, header.Filename)
		if err := cs.fileService.SaveCertificateFromForm(r, "serverCertFile", serverCertPath); err != nil {
			cs.logger.Error(fmt.Sprintf("Failed to save server cert: %v", err))
			return err
		}
		cs.logger.Info(fmt.Sprintf("Saved server certificate to %s", serverCertPath))

		if *serverCert == nil {
			*serverCert = &certificates.CertificateDefs{}
		}
		(*serverCert).PublicKey = serverCertPath
	}
	return nil
}

func (cs *CertificateService) handleServerKeyUpdate(r *http.Request, certDir string, serverCert **certificates.CertificateDefs) error {
	if _, header, err := r.FormFile("serverKeyFile"); err == nil {
		serverKeyPath := filepath.Join(certDir, header.Filename)
		if err := cs.fileService.SaveCertificateFromForm(r, "serverKeyFile", serverKeyPath); err != nil {
			cs.logger.Error(fmt.Sprintf("Failed to save server key: %v", err))
			return err
		}
		cs.logger.Info(fmt.Sprintf("Saved server key to %s", serverKeyPath))

		if *serverCert == nil {
			*serverCert = &certificates.CertificateDefs{}
		}
		(*serverCert).PrivateKey = serverKeyPath
	}
	return nil
}

func (cs *CertificateService) handleClientCertificateUpdate(r *http.Request, certDir string, clientCert **certificates.CertificateDefs) error {
	if _, header, err := r.FormFile("clientCertFile"); err == nil {
		clientCertPath := filepath.Join(certDir, header.Filename)
		if err := cs.fileService.SaveCertificateFromForm(r, "clientCertFile", clientCertPath); err != nil {
			cs.logger.Error(fmt.Sprintf("Failed to save client cert: %v", err))
			return err
		}
		cs.logger.Info(fmt.Sprintf("Saved client CA certificate to %s", clientCertPath))

		if *clientCert == nil {
			*clientCert = &certificates.CertificateDefs{}
		}
		(*clientCert).CaPem = []string{clientCertPath}
	}
	return nil
}
