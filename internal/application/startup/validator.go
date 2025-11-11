package startup

import (
	"errors"

	"github.com/janmbaco/go-infrastructure/v2/configuration"
	ioc "github.com/janmbaco/go-infrastructure/v2/configuration/fileconfig/ioc"
	fileConfigResolver "github.com/janmbaco/go-infrastructure/v2/configuration/fileconfig/ioc/resolver"
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	diskIoc "github.com/janmbaco/go-infrastructure/v2/disk/ioc"
	errorsIoc "github.com/janmbaco/go-infrastructure/v2/errors/ioc"
	eventsIoc "github.com/janmbaco/go-infrastructure/v2/eventsmanager/ioc"
	logsIoc "github.com/janmbaco/go-infrastructure/v2/logs/ioc"
	serverIoc "github.com/janmbaco/go-infrastructure/v2/server/ioc"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	presentationIoc "github.com/janmbaco/go-reverseproxy-ssl/v3/internal/presentation/ioc"
)

type ConfigValidator struct{}

func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

func (cv *ConfigValidator) Validate(configFile string, defaultConfig *domain.Config) error {
	container := dependencyinjection.NewBuilder().
		AddModule(logsIoc.NewLogsModule()).
		AddModule(errorsIoc.NewErrorsModule()).
		AddModule(eventsIoc.NewEventsModule()).
		AddModule(diskIoc.NewDiskModule()).
		AddModule(ioc.NewConfigurationModule()).
		AddModule(serverIoc.NewServerModule()).
		AddModule(presentationIoc.NewPresentationModule()).
		MustBuild()

	configHandler := fileConfigResolver.GetFileConfigHandler(
		container.Resolver(),
		configFile,
		defaultConfig,
	)

	dependencyinjection.RegisterSingleton(container.Register(), func() configuration.ConfigHandler { return configHandler })
	config := configHandler.GetConfig().(*domain.Config)
	return config.Validate()
}

func (cv *ConfigValidator) ValidateRuntime(config interface{}) (bool, error) {
	domainConfig, ok := config.(*domain.Config)
	if !ok {
		return false, errors.New("invalid config type")
	}

	if err := domainConfig.Validate(); err != nil {
		return false, err
	}

	return true, nil
}
