package presentation

import (
	"net/http"
	"strconv"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
)

// FormHandler implementa la responsabilidad de procesar formularios HTTP
type FormHandler struct{}

// NewFormHandler crea una nueva instancia de FormHandler
func NewFormHandler() IFormHandler {
	return &FormHandler{}
}

// ParseVirtualHostForm implementa IFormHandler.ParseVirtualHostForm
func (fh *FormHandler) ParseVirtualHostForm(r *http.Request, defaultVH *domain.WebVirtualHost) (string, uint, error) {
	from := r.FormValue("from")
	if from == "" && defaultVH != nil {
		from = defaultVH.From
	}

	port := uint(8080)
	if defaultVH != nil {
		port = defaultVH.Port
	}

	portStr := r.FormValue("port")
	if portStr != "" {
		if p, err := strconv.ParseUint(portStr, 10, 32); err == nil {
			port = uint(p)
		}
	}

	return from, port, nil
}
