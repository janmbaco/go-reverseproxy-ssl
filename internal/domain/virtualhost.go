package domain

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/google/uuid"
	certs "github.com/janmbaco/go-reverseproxy-ssl/internal/infrastructure/certificates"
)

// VirtualHostBase is used to configure a virtual host.
type VirtualHostBase struct {
	ID                string                 `json:"id,omitempty"`
	From              string                 `json:"from"`
	Scheme            string                 `json:"scheme"`
	HostName          string                 `json:"host_name"`
	Port              uint                   `json:"port"`
	Path              string                 `json:"path"`
	ServerCertificate *certs.CertificateDefs `json:"server_certificate"`
	urlToReplace      string
	pathToDelete      string
	hostToReplace     string
	logger            Logger
}

// EnsureID ensures the virtual host has a unique ID
func (virtualHost *VirtualHostBase) EnsureID() {
	if virtualHost.ID == "" {
		virtualHost.ID = uuid.New().String()
	}
}

// GetID obtains the unique identifier
func (virtualHost *VirtualHostBase) GetID() string {
	return virtualHost.ID
}

// GetFrom obtains the original host
func (virtualHost *VirtualHostBase) GetFrom() string {
	return virtualHost.From
}

// SetURLToReplace  sets the url that replace to the virtual host.
func (virtualHost *VirtualHostBase) SetURLToReplace() {
	virtualHost.urlToReplace = virtualHost.From
	if !strings.HasSuffix(virtualHost.urlToReplace, "/") {
		virtualHost.urlToReplace += "/"
	}
	paths := strings.SplitAfterN(virtualHost.From, "/", 2)
	virtualHost.hostToReplace = strings.ReplaceAll(paths[0], "/", "")
	if len(paths) == 2 && len(paths[1]) > 0 {
		var b strings.Builder
		b.WriteString("/")
		b.WriteString(paths[1])
		if !strings.HasSuffix(paths[1], "/") {
			b.WriteString("/")
		}
		virtualHost.pathToDelete = b.String()
	}
}

// GetHostToReplace gets the host to replace by de virtual host.
func (virtualHost *VirtualHostBase) GetHostToReplace() string {
	return virtualHost.hostToReplace
}

// GetURLToReplace gets the url to replace by de virtual host.
func (virtualHost *VirtualHostBase) GetURLToReplace() string {
	return virtualHost.urlToReplace
}

// GetURL gets the url of the virtual host.
func (virtualHost *VirtualHostBase) GetURL() string {
	return fmt.Sprintf("'%v://%v:%v/%v'", virtualHost.Scheme, virtualHost.HostName, virtualHost.Port, virtualHost.Path)
}

// GetServerCertificate gets de certificate of the server.
func (virtualHost *VirtualHostBase) GetServerCertificate() CertificateProvider {
	return virtualHost.ServerCertificate
}

// GetAuthorizedCAs gets the certificate authorities of the virtual host.
func (virtualHost *VirtualHostBase) GetAuthorizedCAs() []string {
	CAs := make([]string, 0)
	if virtualHost.ServerCertificate != nil && len(virtualHost.ServerCertificate.GetAuthorizedCAs()) > 0 {
		CAs = append(CAs, virtualHost.ServerCertificate.GetAuthorizedCAs()...)
	}
	return CAs
}

// GetHostName gets the host name
func (virtualHost *VirtualHostBase) GetHostName() string {
	var b strings.Builder
	b.WriteString(virtualHost.HostName)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(int(virtualHost.Port)))
	return b.String()
}

func (virtualHost *VirtualHostBase) serve(rw http.ResponseWriter, req *http.Request, directorFunc func(outReq *http.Request), transport http.RoundTripper) {
	(&httputil.ReverseProxy{
		Director:  directorFunc,
		ErrorLog:  virtualHost.logger.GetErrorLogger(),
		Transport: transport,
	}).ServeHTTP(rw, req)
}

func (virtualHost *VirtualHostBase) getPath(virtualPath string) string {
	// Remove pathToDelete from virtualPath
	remainingPath := strings.Replace(virtualPath, virtualHost.pathToDelete, "", 1)

	// Normalize double slashes in remainingPath first
	remainingPath = strings.ReplaceAll(remainingPath, "//", "/")

	var b strings.Builder
	b.WriteString("/")
	if len(virtualHost.Path) > 0 {
		b.WriteString(virtualHost.Path)
		if !strings.HasSuffix(virtualHost.Path, "/") {
			b.WriteString("/")
		}
	}

	// Remove leading slash from remainingPath if Path is empty to avoid double slash
	if len(virtualHost.Path) == 0 && strings.HasPrefix(remainingPath, "/") {
		remainingPath = remainingPath[1:]
	}

	b.WriteString(remainingPath)

	// Final normalization
	result := strings.ReplaceAll(b.String(), "//", "/")

	// Ensure result starts with "/"
	if !strings.HasPrefix(result, "/") {
		result = "/" + result
	}

	return result
}

func (virtualHost *VirtualHostBase) redirectRequest(outReq *http.Request, req *http.Request, setXForwaredHeader bool) {
	outReq.URL.Scheme = virtualHost.Scheme
	outReq.URL.Host = virtualHost.GetHostName()
	outReq.URL.Path = virtualHost.getPath(req.URL.Path)
	outReq.URL.RawQuery = req.URL.RawQuery
	outReq.Header = req.Header
	if setXForwaredHeader {
		outReq.Header.Set("X-Forwarded-Proto", "https")
	}
	virtualHost.logger.Info(fmt.Sprintf("from '%v%v%v' to '%v%v%v'", req.URL.Host, req.URL.Path, req.URL.RawQuery, outReq.URL.Host, outReq.URL.Path, outReq.URL.RawQuery))
}
