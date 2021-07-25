package hosts

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"

	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
)

// IVirtualHost is the definition of a object that represents a Virtual Host to reverse proxy.
type IVirtualHost interface {
	GetFrom() string
	SetUrlToReplace()
	GetHostToReplace() string
	GetUrlToReplace() string
	GetUrl() string
	GetAuthorizedCAs() []string
	GetServerCertificate() *certs.CertificateDefs
	ServeHTTP(rw http.ResponseWriter, req *http.Request)
}

// VirtualHost is used to configure a virtual host.
type VirtualHost struct {
	From              string                 `json:"from"`
	Scheme            string                 `json:"scheme"`
	HostName          string                 `json:"host_name"`
	Port              uint                   `json:"port"`
	Path              string                 `json:"path"`
	ServerCertificate *certs.CertificateDefs `json:"server_certificate"`
	urlToReplace      string
	pathToDelete      string
	hostToReplace     string
}

// GetFrom obtains the original host
func (virtualHost *VirtualHost) GetFrom() string {
	return virtualHost.From
}

// SetUrlToReplace sets the url that replace to the virtual host.
func (virtualHost *VirtualHost) SetUrlToReplace() {
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
func (virtualHost *VirtualHost) GetHostToReplace() string {
	return virtualHost.hostToReplace
}

// GetUrlToReplace gets the url to replace by de virtual host.
func (virtualHost *VirtualHost) GetUrlToReplace() string {
	return virtualHost.urlToReplace
}

// GetUrl gets the url of the virtual host.
func (virtualHost *VirtualHost) GetUrl() string {
	return fmt.Sprintf("'%v://%v:%v/%v'", virtualHost.Scheme, virtualHost.HostName, virtualHost.Port, virtualHost.Path)
}

// GetServerCertificate gets de certificate of the server.
func (virtualHost *VirtualHost) GetServerCertificate() *certs.CertificateDefs {
	return virtualHost.ServerCertificate
}

// GetAuthorizedCAs gets the certificate authorities of the virtual host.
func (virtualHost *VirtualHost) GetAuthorizedCAs() []string {
	CAs := make([]string, 0)
	if virtualHost.ServerCertificate != nil && len(virtualHost.ServerCertificate.CaPem) > 0 {
		CAs = append(CAs, virtualHost.ServerCertificate.CaPem)
	}
	return CAs
}

func (virtualHost *VirtualHost) serve(rw http.ResponseWriter, req *http.Request, directorFunc func(outReq *http.Request), transport http.RoundTripper) {
	(&httputil.ReverseProxy{
		Director:  directorFunc,
		ErrorLog:  logs.Log.ErrorLogger,
		Transport: transport,
	}).ServeHTTP(rw, req)
}

func (virtualHost *VirtualHost) getPath(virtualPath string) string {
	var b strings.Builder
	b.WriteString("/")
	if len(virtualHost.Path) > 0 {
		b.WriteString(virtualHost.Path)
		if !strings.HasSuffix(virtualHost.Path, "/") {
			b.WriteString("/")
		}
	}
	b.WriteString(strings.Replace(virtualPath, virtualHost.pathToDelete, "", 1))
	return strings.ReplaceAll(b.String(), "//", "/")
}

func (virtualHost *VirtualHost) redirectRequest(outReq *http.Request, req *http.Request) {
	outReq.URL.Scheme = virtualHost.Scheme
	outReq.URL.Host = virtualHost.getHost()
	outReq.URL.Path = virtualHost.getPath(req.URL.Path)
	outReq.URL.RawQuery = req.URL.RawQuery
	outReq.Header = req.Header
	outReq.Header.Set("X-Forwarded-Proto", "https")

	logs.Log.Info(fmt.Sprintf("from '%v%v%v' to '%v%v%v'", req.URL.Host, req.URL.Path, req.URL.RawQuery, outReq.URL.Host, outReq.URL.Path, outReq.URL.RawQuery))
}

func (virtualHost *VirtualHost) getHost() string {
	var b strings.Builder
	b.WriteString(virtualHost.HostName)
	b.WriteString(":")
	b.WriteString(strconv.Itoa(int(virtualHost.Port)))
	return b.String()
}
