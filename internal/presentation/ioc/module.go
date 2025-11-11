package ioc

import (
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/presentation"
)

// PresentationModule implementa Module para servicios de presentación
type PresentationModule struct{}

// NewPresentationModule crea un nuevo módulo de presentación
func NewPresentationModule() *PresentationModule {
	return &PresentationModule{}
}

// RegisterServices registra todos los servicios de presentación
func (m *PresentationModule) RegisterServices(register dependencyinjection.Register) error {
	// FormHandler - sin parámetros
	dependencyinjection.RegisterSingleton[presentation.IFormHandler](
		register,
		presentation.NewFormHandler,
	)

	// FileService - necesita logger
	dependencyinjection.RegisterSingletonWithParams[presentation.IFileService](
		register,
		presentation.NewFileService,
		map[int]string{
			0: "logger",
		},
	)

	// CertificateService - necesita logger (fileService se resuelve automáticamente)
	dependencyinjection.RegisterSingletonWithParams[presentation.ICertificateService](
		register,
		presentation.NewCertificateService,
		map[int]string{
			1: "logger",
		},
	)

	// VirtualHostService - necesita logger (otras dependencias automáticas)
	dependencyinjection.RegisterSingletonWithParams[presentation.IVirtualHostService](
		register,
		presentation.NewVirtualHostService,
		map[int]string{
			4: "logger",
		},
	)

	// ConfigUI - necesita configHandler y logger
	dependencyinjection.RegisterSingletonWithParams[*presentation.ConfigUI](
		register,
		presentation.NewConfigUI,
		map[int]string{
			0: "configHandler",
			3: "logger",
		},
	)

	return nil
}
