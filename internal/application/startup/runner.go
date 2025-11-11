package startup

import (
	"fmt"
	"os"

	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	applicationResolver "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application/ioc/resolver"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	infrastructureResolver "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/ioc/resolver"
	presentationResolver "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/presentation/ioc/resolver"

	fileConfigResolver "github.com/janmbaco/go-infrastructure/v2/configuration/fileconfig/ioc/resolver"
	logsResolver "github.com/janmbaco/go-infrastructure/v2/logs/ioc/resolver"
	serverResolver "github.com/janmbaco/go-infrastructure/v2/server/ioc/resolver"
)

type ApplicationRunner struct {
	configFile    string
	defaultConfig *domain.Config
	validateFunc  func(interface{}) (bool, error)
}

func NewApplicationRunner(configFile string, defaultConfig *domain.Config, validateFunc func(interface{}) (bool, error)) *ApplicationRunner {
	return &ApplicationRunner{
		configFile:    configFile,
		defaultConfig: defaultConfig,
		validateFunc:  validateFunc,
	}
}

func (ar *ApplicationRunner) Start(container dependencyinjection.Container) {
	configHandler := fileConfigResolver.GetFileConfigHandler(
		container.Resolver(),
		ar.configFile,
		ar.defaultConfig,
	)

	logger := logsResolver.GetLogger(container.Resolver())

	httpRedirector := infrastructureResolver.GetHTTPRedirector(container.Resolver(), logger)
	proxyConfigurator := applicationResolver.GetReverseProxyConfigurator(container.Resolver(), logger, configHandler)

	ar.setLogConfiguration(configHandler.GetConfig().(*domain.Config), logger)

	// Start HTTP redirector in background
	go func() {
		if err := httpRedirector.Start(); err != nil {
			logger.Error(fmt.Sprintf("HTTP redirector error: %v", err))
		}
	}()

	// Start ConfigUI in background (with error handling)
	configUI := presentationResolver.GetConfigUI(container.Resolver(), configHandler, logger)
	if configUI != nil {
		go func() {
			if err := configUI.Start(configHandler.GetConfig().(*domain.Config).ConfigUIPort); err != nil {
				logger.Error(fmt.Sprintf("ConfigUI server error: %v", err))
			}
		}()
	} else {
		logger.Error("Failed to initialize ConfigUI - templates may be corrupted. ConfigUI will not be available.")
	}

	listenerBuilder := serverResolver.GetListenerBuilder(container.Resolver(), configHandler).
		SetBootstrapper(proxyConfigurator.Configure).
		SetConfigValidatorFunc(ar.validateFunc)

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

func (ar *ApplicationRunner) setLogConfiguration(config *domain.Config, logger domain.Logger) {
	logger.SetDir(config.LogsDir)
	logger.SetConsoleLevel(config.LogConsoleLevel)
	logger.SetFileLogLevel(config.LogFileLevel)
}
