package resolver

import (
	"github.com/janmbaco/go-infrastructure/v2/configuration"
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/presentation"
)

func GetConfigUI(resolver dependencyinjection.Resolver, configHandler configuration.ConfigHandler, logger domain.Logger) *presentation.ConfigUI {
	result := resolver.Type(new(*presentation.ConfigUI), map[string]interface{}{
		"configHandler": configHandler,
		"logger":        logger,
	})
	if configUI, ok := result.(*presentation.ConfigUI); ok {
		return configUI
	}
	panic("failed to resolve ConfigUI")
}
