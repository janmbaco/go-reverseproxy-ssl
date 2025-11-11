package resolver

import (
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
)

// GetHTTPRedirector obtiene el HTTPRedirector del contenedor IoC
func GetHTTPRedirector(resolver dependencyinjection.Resolver, logger domain.Logger) domain.HTTPRedirector {
	return dependencyinjection.ResolveWithParams[domain.HTTPRedirector](resolver, map[string]any{"logger": logger})
}
