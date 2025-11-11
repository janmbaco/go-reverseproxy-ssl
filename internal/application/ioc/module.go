package ioc

import (
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application/hosts_resolver"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
)

// ApplicationModule implementa Module para servicios de aplicación
type ApplicationModule struct{}

// NewApplicationModule crea un nuevo módulo de aplicación
func NewApplicationModule() *ApplicationModule {
	return &ApplicationModule{}
}

// RegisterServices registra todos los servicios de aplicación
func (m *ApplicationModule) RegisterServices(register dependencyinjection.Register) error {
	// Registrar VirtualHostResolver como singleton con resolución automática de dependencias
	dependencyinjection.RegisterSingletonWithParams[domain.VirtualHostResolver](
		register,
		hosts_resolver.NewVirtualHostResolver,
		map[int]string{1: "logger"},
	)

	// Registrar ServerState como singleton
	dependencyinjection.RegisterSingleton(
		register,
		func() *application.ServerState {
			return &application.ServerState{}
		},
	)

	// Registrar ReverseProxyConfigurator como singleton con resolución automática de dependencias
	dependencyinjection.RegisterSingletonWithParams[*application.ReverseProxyConfigurator](
		register,
		application.NewReverseProxyConfigurator,
		map[int]string{0: "logger", 2: "configHandler"},
	)

	return nil
}
