package infrastructure

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
)

type HTTPRedirector struct {
	mux           *http.ServeMux
	redirectRules map[string]string
	server        *http.Server
	logger        domain.Logger
}

func NewHTTPRedirector(logger domain.Logger) *HTTPRedirector {
	hr := &HTTPRedirector{
		mux:           http.NewServeMux(),
		redirectRules: make(map[string]string),
		logger:        logger,
	}

	hr.server = &http.Server{
		Addr:    ":80",
		Handler: hr.mux,
	}

	hr.mux.HandleFunc("/", hr.handleRedirect)

	return hr
}

func (hr *HTTPRedirector) handleRedirect(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	path := r.URL.Path

	if httpsURL, exists := hr.redirectRules[host]; exists {
		hr.logger.Info(fmt.Sprintf("Redirecting HTTP %s%s → %s", host, path, httpsURL))
		http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
		return
	}

	httpsURL := fmt.Sprintf("https://%s%s", host, r.URL.RequestURI())
	hr.logger.Info(fmt.Sprintf("Generic redirect HTTP %s%s → %s", host, path, httpsURL))
	http.Redirect(w, r, httpsURL, http.StatusMovedPermanently)
}

func (hr *HTTPRedirector) UpdateRedirectRules(virtualHosts []domain.IVirtualHost) {
	newRules := make(map[string]string)

	for _, vh := range virtualHosts {
		from := vh.GetFrom()

		httpsURL := fmt.Sprintf("https://%s", from)
		if !strings.HasPrefix(from, "http") {
			httpsURL = fmt.Sprintf("https://%s", from)
		}

		newRules[from] = httpsURL
		hr.logger.Info(fmt.Sprintf("Added redirect rule: %s → %s", from, httpsURL))
	}

	hr.redirectRules = newRules
	hr.logger.Info(fmt.Sprintf("Updated %d redirect rules", len(newRules)))
}

func (hr *HTTPRedirector) Start() error {
	hr.logger.Info("Starting HTTP redirector on port 80")
	return hr.server.ListenAndServe()
}

func (hr *HTTPRedirector) Stop() error {
	hr.logger.Info("Stopping HTTP redirector")
	return hr.server.Close()
}
