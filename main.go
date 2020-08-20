package main

import (
	"flag"
	"fmt"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/janmbaco/go-infrastructure/config"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-infrastructure/server"
	"github.com/janmbaco/go-reverseproxy-ssl/configs"
	"github.com/russross/blackfriday/v2"
)

var Config *configs.Config

func main() {

	var configfile = flag.String("config", os.Args[0]+".config", "Config File")
	flag.Parse()

	Config = setDefaultConfig()
	configHandler := config.NewFileConfigHandler(*configfile)
	configHandler.Load(Config)
	setLogConfiguration()
	configHandler.OnModifiedConfigSubscriber(setLogConfiguration)

	logs.Log.Info("")
	logs.Log.Info("Start Server Application")
	logs.Log.Info("")

	//redirect http to https
	go func() {
		server.NewListener(configHandler, redirectHttpToHttps).Start()
	}()
	//start server
	server.NewListener(configHandler, reverseProxy).Start()
}

func setLogConfiguration() {
	logs.Log.SetDir(Config.LogsDir)
	logs.Log.SetConsoleLevel(Config.LogConsoleLevel)
	logs.Log.SetFileLogLevel(Config.LogFileLevel)
}

func setDefaultConfig() *configs.Config {
	//default config if file is not found
	return &configs.Config{
		VirtualHost: map[string]*configs.VirtualHost{
			"example.host.com": {
				Scheme:   "http",
				HostName: "localhost",
				Port:     2256,
			},
		},
		DefaultHost:      "localhost",
		ReverseProxyPort: ":443",
		LogConsoleLevel:  logs.Trace,
		LogFileLevel:     logs.Warning,
		LogsDir:          "./logs",
	}
}

func redirectHttpToHttps(serverSetter *server.ServerSetter) {
	logs.Log.Info("Start Redirect Server from http to https")
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	}))
	serverSetter.Addr = ":80"
	serverSetter.Handler = mux
}

func reverseProxy(serverSetter *server.ServerSetter) {

	mux := http.NewServeMux()

	var virtualHost []string
	var caPems []string
	var isRegisteredDefaultHost bool

	for name, vHost := range Config.VirtualHost {

		virtualHost = append(virtualHost, name)

		if len(vHost.CaPem) > 0 {
			caPems = append(caPems, vHost.CaPem)
		}

		if name == Config.DefaultHost {
			isRegisteredDefaultHost = true
		}

		logs.Log.Info(fmt.Sprintf("register proxy from: '%v' to '%v://%v:%v'", name, vHost.Scheme, vHost.HostName, vHost.Port))
		mux.Handle(name+"/", vHost)
		redirectToWWW(name, mux)
	}

	if !isRegisteredDefaultHost {
		logs.Log.Info(fmt.Sprintf("register default host: '%v'", Config.DefaultHost))
		mux.HandleFunc(Config.DefaultHost+"/", func(w http.ResponseWriter, r *http.Request) {
			if tmpl, err := template.ParseFiles("./html/index.html"); err != nil {
				errorhandler.TryPanic(err)
			} else {
				if markDown, err := ioutil.ReadFile("README.md"); err != nil {
					errorhandler.TryPanic(err)
				} else {
					type pipe struct {
						Home        template.HTML
						VirtualHost map[string]*configs.VirtualHost
					}
					home := strings.SplitN(string(markDown), "\n", 2)[1]
					errorhandler.TryPanic(tmpl.ExecuteTemplate(w, "index.html", &pipe{template.HTML(blackfriday.Run([]byte(home))), Config.VirtualHost}))
				}

			}
		})
		virtualHost = append(virtualHost, Config.DefaultHost)

		if Config.DefaultHost != "localhost" {
			redirectToWWW(Config.DefaultHost, mux)
		}
	}

	serverSetter.Addr = Config.ReverseProxyPort
	serverSetter.Handler = mux

	if Config.DefaultHost != "localhost" {
		serverSetter.TLSConfig = configs.GetTlsConfig(virtualHost, caPems)
	}
}

func redirectToWWW(hostname string, mux *http.ServeMux) {
	if strings.Contains(hostname, "www") {
		//redirect to web host with www.
		mux.Handle(strings.Replace(hostname, "www.", "", 1)+"/",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+hostname, http.StatusMovedPermanently)
			}))
	}
}
