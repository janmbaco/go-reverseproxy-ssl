package infrastructure

import (
	"errors"
	"fmt"
	"os"

	"github.com/janmbaco/go-infrastructure/v2/configuration/fileconfig/ioc"
	fileConfigResolver "github.com/janmbaco/go-infrastructure/v2/configuration/fileconfig/ioc/resolver"
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	diskIoc "github.com/janmbaco/go-infrastructure/v2/disk/ioc"
	errorsIoc "github.com/janmbaco/go-infrastructure/v2/errors/ioc"
	eventsIoc "github.com/janmbaco/go-infrastructure/v2/eventsmanager/ioc"
	"github.com/janmbaco/go-infrastructure/v2/logs"
	logsIoc "github.com/janmbaco/go-infrastructure/v2/logs/ioc"
	logsResolver "github.com/janmbaco/go-infrastructure/v2/logs/ioc/resolver"
	serverIoc "github.com/janmbaco/go-infrastructure/v2/server/ioc"
	serverResolver "github.com/janmbaco/go-infrastructure/v2/server/ioc/resolver"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application"
	resolver "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application/hosts_resolver"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/presentation"
)

type ServerBootstrapper struct {
	configFile string
}

func NewServerBootstrapper(configFile string) *ServerBootstrapper {
	return &ServerBootstrapper{configFile: configFile}
}

func (sb *ServerBootstrapper) Start() {
	container := dependencyinjection.NewBuilder().
		AddModule(logsIoc.NewLogsModule()).
		AddModule(errorsIoc.NewErrorsModule()).
		AddModule(eventsIoc.NewEventsModule()).
		AddModule(diskIoc.NewDiskModule()).
		AddModule(ioc.NewConfigurationModule()).
		AddModule(serverIoc.NewServerModule()).
		MustBuild()

	configHandler := fileConfigResolver.GetFileConfigHandler(
		container.Resolver(),
		sb.configFile,
		sb.createDefaultConfig(),
	)

	logger := logsResolver.GetLogger(container.Resolver())
	serverState := &application.ServerState{}

	vhResolver := resolver.NewVirtualHostResolver(container, logger)
	httpRedirector := NewHTTPRedirector(logger)
	proxyConfigurator := application.NewReverseProxyConfigurator(logger, vhResolver, configHandler, serverState)

	sb.setLogConfiguration(configHandler.GetConfig().(*domain.Config), logger)

	// Start HTTP redirector in background
	go func() {
		if err := httpRedirector.Start(); err != nil {
			logger.Error(fmt.Sprintf("HTTP redirector error: %v", err))
		}
	}()

	// Start ConfigUI in background (with error handling)
	configUI := presentation.NewConfigUI(configHandler, vhResolver, logger)
	if configUI != nil {
		go configUI.Start(configHandler.GetConfig().(*domain.Config).ConfigUIPort)
	} else {
		logger.Error("Failed to initialize ConfigUI - templates may be corrupted. ConfigUI will not be available.")
	}

	listenerBuilder := serverResolver.GetListenerBuilder(container.Resolver(), configHandler).
		SetBootstrapper(proxyConfigurator.Configure).
		SetConfigValidatorFunc(sb.validateConfig)

	listener, err := listenerBuilder.GetListener()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create listener: %v\n", err)
		os.Exit(1)
	}

	finish := listener.Start()

	if err := <-finish; err != nil {
		logger.Error(fmt.Sprintf("Server error: %v", err))
		os.Exit(1)
	}
}

// ValidateConfig loads and validates the configuration without starting the server
func (sb *ServerBootstrapper) ValidateConfig() error {
	container := dependencyinjection.NewBuilder().
		AddModule(logsIoc.NewLogsModule()).
		AddModule(errorsIoc.NewErrorsModule()).
		AddModule(eventsIoc.NewEventsModule()).
		AddModule(diskIoc.NewDiskModule()).
		AddModule(ioc.NewConfigurationModule()).
		AddModule(serverIoc.NewServerModule()).
		MustBuild()

	configHandler := fileConfigResolver.GetFileConfigHandler(
		container.Resolver(),
		sb.configFile,
		sb.createDefaultConfig(),
	)

	config := configHandler.GetConfig().(*domain.Config)
	return config.Validate()
}

// validateConfig validates configuration for runtime changes (used by listener)
// Returns (isValid, error) - isValid is true if validation passes, false otherwise
func (sb *ServerBootstrapper) validateConfig(config interface{}) (bool, error) {
	domainConfig, ok := config.(*domain.Config)
	if !ok {
		return false, errors.New("invalid config type")
	}

	if err := domainConfig.Validate(); err != nil {
		return false, err
	}

	return true, nil
}

func (sb *ServerBootstrapper) createDefaultConfig() *domain.Config {
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

func (sb *ServerBootstrapper) setLogConfiguration(config *domain.Config, logger domain.Logger) {
	logger.SetDir(config.LogsDir)
	logger.SetConsoleLevel(config.LogConsoleLevel)
	logger.SetFileLogLevel(config.LogFileLevel)
}
