package application

import (
	"net/http"
	"strings"
)

// RedirectToWWW sets up a redirect handler for non-www domains to their www equivalent
func RedirectToWWW(hostname string, mux *http.ServeMux) {
	if strings.HasPrefix(hostname, "www") {
		mux.Handle(strings.Replace(hostname, "www.", "", 1)+"/",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+hostname+r.RequestURI, http.StatusMovedPermanently)
			}))
	}
}
