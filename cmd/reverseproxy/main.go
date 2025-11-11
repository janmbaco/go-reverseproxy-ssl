package main

import (
	"flag"
	"os"

	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/application/startup"
)

func main() {
	var configFile = flag.String("config", "", "globalConfig File")
	var validateOnly = flag.Bool("validate", false, "validate configuration and exit")
	flag.Parse()
	if len(*configFile) == 0 {
		_, _ = os.Stderr.WriteString("You must set a config file!\n")
		flag.PrintDefaults()
		return
	}

	bootstrapper := NewServerBootstrapper()
	defaultConfig := bootstrapper.CreateDefaultConfig()

	if *validateOnly {
		validator := startup.NewConfigValidator()
		if err := validator.Validate(*configFile, defaultConfig); err != nil {
			_, _ = os.Stderr.WriteString("Configuration validation failed: " + err.Error() + "\n")
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString("Configuration is valid\n")
		return
	}

	container := bootstrapper.BuildContainer()

	validator := startup.NewConfigValidator()
	runner := startup.NewApplicationRunner(*configFile, defaultConfig, validator.ValidateRuntime)

	runner.Start(container)
}
