package domain

import (
	"crypto/tls"
	"log"
	"net/http"

	"github.com/janmbaco/go-infrastructure/v2/logs"
)

// Logger interface for logging
type Logger interface {
	Info(msg string)
	Error(msg string)
	SetDir(dir string)
	SetConsoleLevel(level logs.LogLevel)
	SetFileLogLevel(level logs.LogLevel)
	GetErrorLogger() *log.Logger
	PrintError(level logs.LogLevel, err error)
}

// CertificateManager interface for managing certificates
type CertificateManager interface {
	AddCertificate(host string, cert *tls.Certificate)
	AddAutoCertificate(host string)
	HasCertificateFor(host string) bool
	GetTLSConfig() *tls.Config
	AddClientCA(cas []string)
}

// CertificateProvider interface for certificate definitions
type CertificateProvider interface {
	GetCertificate() (*tls.Certificate, error)
	GetTLSConfig() (*tls.Config, error)
	GetAuthorizedCAs() []string
}

// IVirtualHost is the definition of an object that represents a Virtual Host to reverse proxy.
type IVirtualHost interface {
	http.Handler
	GetID() string
	GetFrom() string
	SetURLToReplace()
	GetHostToReplace() string
	GetURLToReplace() string
	GetURL() string
	GetAuthorizedCAs() []string
	GetServerCertificate() CertificateProvider
	GetHostName() string
	EnsureID()
}

// VirtualHostResolver interface for resolving virtual hosts
type VirtualHostResolver interface {
	Resolve(config *Config) ([]IVirtualHost, error)
}

// HTTPRedirector interface for managing HTTP to HTTPS redirects
type HTTPRedirector interface {
	UpdateRedirectRules(hosts []IVirtualHost)
	Start() error
}

// Tenants for virtual hosts
const (
	WebVirtualHostTenant      = "WebVirtualHost"
	SSHVirtualHostTenant      = "SshVirtualHost"
	GrpcJSONVirtualHostTenant = "GrpcJSONVirtualHost"
	GrpcVirtualHostTenant     = "GrpcVirtualHost"
	GrpcWebVirtualHostTenant  = "GrpcWebVirtualHost"
)
