package main

import (
	"github.com/janmbaco/go-infrastructure/v2/configuration/fileconfig/ioc"
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	diskIoc "github.com/janmbaco/go-infrastructure/v2/disk/ioc"
	errorsIoc "github.com/janmbaco/go-infrastructure/v2/errors/ioc"
	eventsIoc "github.com/janmbaco/go-infrastructure/v2/eventsmanager/ioc"
	"github.com/janmbaco/go-infrastructure/v2/logs"
	logsIoc "github.com/janmbaco/go-infrastructure/v2/logs/ioc"
	serverIoc "github.com/janmbaco/go-infrastructure/v2/server/ioc"
	applicationIoc "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application/ioc"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	infrastructureIoc "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/ioc"
	presentationIoc "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/presentation/ioc"
)

type ServerBootstrapper struct{}

func NewServerBootstrapper() *ServerBootstrapper {
	return &ServerBootstrapper{}
}

func (sb *ServerBootstrapper) BuildContainer() dependencyinjection.Container {
	return dependencyinjection.NewBuilder().
		AddModule(logsIoc.NewLogsModule()).
		AddModule(errorsIoc.NewErrorsModule()).
		AddModule(eventsIoc.NewEventsModule()).
		AddModule(diskIoc.NewDiskModule()).
		AddModule(ioc.NewConfigurationModule()).
		AddModule(infrastructureIoc.NewInfrastructureModule()).
		AddModule(applicationIoc.NewApplicationModule()).
		AddModule(serverIoc.NewServerModule()).
		AddModule(presentationIoc.NewPresentationModule()).
		MustBuild()
}

func (sb *ServerBootstrapper) CreateDefaultConfig() *domain.Config {
	return &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "www.example.com",
						Scheme:   "http",
						HostName: "localhost",
						Port:     8080,
					},
				},
			},
		},
		DefaultHost:      "localhost",
		ReverseProxyPort: ":443",
		LogConsoleLevel:  logs.Trace,
		LogFileLevel:     logs.Trace,
		LogsDir:          "./logger",
		ConfigUIPort:     ":8081",
	}
}
