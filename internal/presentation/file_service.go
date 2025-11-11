package presentation

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
)

// FileService implementa la responsabilidad de gestionar operaciones de archivo
type FileService struct {
	logger domain.Logger
}

// NewFileService crea una nueva instancia de FileService
func NewFileService(logger domain.Logger) IFileService {
	return &FileService{
		logger: logger,
	}
}

// CreateCertDirectory implementa IFileService.CreateCertDirectory
func (fs *FileService) CreateCertDirectory(config *domain.Config, from string) (string, error) {
	// Get cert base directory from config or use default
	certBaseDir := config.CertDir
	if certBaseDir == "" {
		certBaseDir = "/app/certs"
	}

	// Create cert directory based on 'from' (sanitize for filesystem)
	safeName := fs.sanitizePathName(from)
	certDir := filepath.Join(certBaseDir, safeName)

	if err := os.MkdirAll(certDir, 0755); err != nil {
		fs.logger.Error(fmt.Sprintf("Failed to create cert directory %s: %v", certDir, err))
		return "", err
	}

	return certDir, nil
}

// SaveCertificateFromForm implementa IFileService.SaveCertificateFromForm
func (fs *FileService) SaveCertificateFromForm(r *http.Request, fieldName, destPath string) error {
	file, _, err := r.FormFile(fieldName)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func() { _ = destFile.Close() }()

	_, err = io.Copy(destFile, file)
	return err
}

// sanitizePathName sanitizes a path name to prevent directory traversal attacks
func (fs *FileService) sanitizePathName(name string) string {
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
