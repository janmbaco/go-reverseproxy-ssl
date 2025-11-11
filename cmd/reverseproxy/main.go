package main

import (
	"flag"
	"os"

	"github.com/janmbaco/go-reverseproxy-ssl/internal/infrastructure"
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

	bootstrapper := infrastructure.NewServerBootstrapper(*configFile)

	if *validateOnly {
		// Only validate configuration
		if err := bootstrapper.ValidateConfig(); err != nil {
			_, _ = os.Stderr.WriteString("Configuration validation failed: " + err.Error() + "\n")
			os.Exit(1)
		}
		_, _ = os.Stdout.WriteString("Configuration is valid\n")
		return
	}

	bootstrapper.Start()
}
