package servers

import (
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/disk"
)

type config struct {
	VirtualHost      map[string]*VirtualHost `json:"virtual_hosts"`
	DefaultHost      string                  `json:"default_host"`
	ReverseProxyPort string                  `json:"reverse_proxy_port"`
	LogConsoleLevel  cross.LogLevel          `json:"log_console_level"`
	LogFileLevel     cross.LogLevel          `json:"log_file_level"`
	LogsDir          string                  `json:"logs_dir"`
}

var Config *config

func init(){
	//default config if not file is found
	Config = &config{
		VirtualHost:  map[string]*VirtualHost{
			"example.host.com" : {
				Scheme: "http",
				HostName: "localhost",
				Port: 2256,
			},
		},
		DefaultHost : "localhost",
		ReverseProxyPort: ":443",
		LogConsoleLevel:  cross.Trace,
		LogFileLevel:     cross.Warning,
	}
	//set config dir
	disk.ConfigFile.SetDir("../configs")
	//set config  Constructor
	disk.ConfigFile.ConstructorContent = func() interface{}{
		return &config{}
	}
	//set Copy Config
	disk.ConfigFile.CopyContent = func(from interface{}, to interface{}) {
		fromConf := from.(*config)
		toConf := to.(*config)
		*toConf = *fromConf
	}

	//write o read config from file
	disk.ConfigFile.Load(Config)
}
