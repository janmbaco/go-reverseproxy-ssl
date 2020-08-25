package hosts

import (
	"fmt"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
)

type IVirtualHost interface {
	SetUrlToReplace(url string)
	GetHostToReplace() string
	GetUrlToReplace() string
	GetUrl() string
	GetCaPem() string
	ServeHTTP(rw http.ResponseWriter, req *http.Request)
}

type VirtualHost struct {
	Scheme        string         `json:"scheme"`
	HostName      string         `json:"host_name"`
	Port          uint           `json:"port"`
	Path          string         `json:"path"`
	TlsDefs       *certs.TlsDefs `json:"tls_config"`
	urlToReplace  string
	pathToDelete  string
	hostToReplace string
}

func (this *VirtualHost) SetUrlToReplace(url string) {
	this.urlToReplace = url
	if !strings.HasSuffix(this.urlToReplace, "/") {
		this.urlToReplace += "/"
	}
	paths := strings.SplitAfterN(url, "/", 2)
	this.hostToReplace = paths[0]
	if len(paths) == 2 && len(paths[1]) > 0 {
		var b strings.Builder
		b.WriteString("/")
		b.WriteString(paths[1])
		if !strings.HasSuffix(paths[1], "/") {
			b.WriteString("/")
		}
		this.pathToDelete = b.String()
	}
}

func (this *VirtualHost) GetHostToReplace() string {
	return this.hostToReplace
}

func (this *VirtualHost) GetUrlToReplace() string {
	return this.urlToReplace
}

func (this *VirtualHost) GetCaPem() string {
	var result string
	if this.TlsDefs != nil && len(this.TlsDefs.CaPem) > 0 {
		result = this.TlsDefs.CaPem
	}
	return result
}

func (this *VirtualHost) GetUrl() string {
	return fmt.Sprintf("'%v://%v:%v/%v'", this.Scheme, this.HostName, this.Port, this.Path)
}

func (this *VirtualHost) serve(rw http.ResponseWriter, req *http.Request, directorFunc func(outReq *http.Request), transport http.RoundTripper) {
	_, err := url.Parse(this.Scheme + "://" + this.HostName + ":" + strconv.Itoa(int(this.Port)))
	errorhandler.TryPanic(err)
	(&httputil.ReverseProxy{
		Director:  directorFunc,
		ErrorLog:  logs.Log.ErrorLogger,
		Transport: transport,
	}).ServeHTTP(rw, req)
}

func (this *VirtualHost) getPath(virtualPath string) string {
	var b strings.Builder
	if len(this.Path) > 0 {
		if !strings.HasPrefix(this.Path, "/") {
			b.WriteString("/")
		}
		b.WriteString(this.Path)
		if !strings.HasSuffix(this.Path, "/") {
			b.WriteString("/")
		}
	}
	b.WriteString(strings.Replace(virtualPath, this.pathToDelete, "", 1))
	return b.String()
}

func (this *VirtualHost) redirectRequest(outReq *http.Request, req *http.Request) {
	outReq.URL.Scheme = this.Scheme
	outReq.URL.Host = this.HostName + ":" + strconv.Itoa(int(this.Port))
	outReq.URL.Path = this.getPath(req.URL.Path)
	outReq.URL.RawQuery = req.URL.RawQuery
	outReq.Header = req.Header
	outReq.Header.Set("X-Forwarded-Proto", "https")

	logs.Log.Info(fmt.Sprintf("from '%v%v%v' to '%v%v%v'", req.URL.Host, req.URL.Path, req.URL.RawQuery, outReq.URL.Host, outReq.URL.Path, outReq.URL.RawPath))
}
