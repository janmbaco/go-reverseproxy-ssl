package ioc

import (
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure"
)

// InfrastructureModule implementa Module para servicios de infraestructura
type InfrastructureModule struct{}

// NewInfrastructureModule crea un nuevo módulo de infraestructura
func NewInfrastructureModule() *InfrastructureModule {
	return &InfrastructureModule{}
}

// RegisterServices registra todos los servicios de infraestructura
func (m *InfrastructureModule) RegisterServices(register dependencyinjection.Register) error {
	// Registrar HTTPRedirector como singleton con resolución automática de dependencias
	dependencyinjection.RegisterSingletonWithParams[domain.HTTPRedirector](
		register,
		infrastructure.NewHTTPRedirector,
		map[int]string{0: "logger"},
	)

	return nil
}
