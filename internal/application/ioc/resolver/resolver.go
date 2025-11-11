package resolver

import (
	"github.com/janmbaco/go-infrastructure/v2/configuration"
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
)

// GetReverseProxyConfigurator obtiene el ReverseProxyConfigurator del contenedor IoC
func GetReverseProxyConfigurator(resolver dependencyinjection.Resolver, logger domain.Logger, configHandler configuration.ConfigHandler) *application.ReverseProxyConfigurator {

	return dependencyinjection.ResolveWithParams[*application.ReverseProxyConfigurator](
		resolver,
		map[string]interface{}{
			"logger":        logger,
			"configHandler": configHandler,
		},
	)
}
