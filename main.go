package main

import (
	"flag"
	"fmt"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/grpcUtil"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/janmbaco/go-infrastructure/config"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-infrastructure/server"
	"github.com/janmbaco/go-reverseproxy-ssl/configs"
)

var Config *configs.Config

type fnSetOptionsVHost func(hosts.IVirtualHost)

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
		GrpcJsonVirtualHosts: map[string]*hosts.GrpcJsonVirtualHost{
			"www.pareceproxy.com": {
				VirtualHost: &hosts.VirtualHost{
					Scheme:   "http",
					HostName: "localhost",
					Port:     8080,
				},
			},
		},
		GrpcWebVirtualHosts: map[string]*hosts.GrpcWebVirtualHost{
			"www.example.com": {
				VirtualHost: &hosts.VirtualHost{
					Scheme:   "http",
					HostName: "loclahost",
					Port:     8080,
				},
				GrpcWebProxy: &grpcUtil.GrpcWebProxy{
					AllowAllOrigins: true,
					UseWebSockets:   true,
				},
			},
		},
		DefaultHost:      "localhost",
		ReverseProxyPort: ":443",
		LogConsoleLevel:  logs.Trace,
		LogFileLevel:     logs.Trace,
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

	registerVirtualHost(mux, Config.WebVirtualHosts, func(vHost hosts.IVirtualHost) {
		virtualHost = append(virtualHost, vHost.GetHostToReplace())
		if caPem := vHost.GetCaPem(); len(caPem) > 0 {
			caPems = append(caPems, caPem)
		}
	})

	registerVirtualHost(mux, Config.SshVirtualHosts, nil)
	registerVirtualHost(mux, Config.GrpcJsonVirtualHosts, nil)
	registerVirtualHost(mux, Config.GrpcWebVirtualHosts, nil)

	logs.Log.Info(fmt.Sprintf("register default host: '%v'", Config.DefaultHost))
	mux.HandleFunc(Config.DefaultHost+"/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("started..."))
	})
	virtualHost = append(virtualHost, Config.DefaultHost)

	if Config.DefaultHost != "localhost" {
		redirectToWWW(Config.DefaultHost, mux)
	}

	serverSetter.Addr = Config.ReverseProxyPort
	serverSetter.Handler = mux
	serverSetter.TLSConfig = certs.GetAutoCertConfig(virtualHost, caPems)

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

func registerVirtualHost(mux *http.ServeMux, virtualHosts interface{}, setVHostOptions fnSetOptionsVHost) {
	for name, vHost := range transformMap(virtualHosts) {
		vHost.SetUrlToReplace(name)
		urlToReplace := vHost.GetUrlToReplace()
		logs.Log.Info(fmt.Sprintf("register proxy from: '%v' to '%v'", name, vHost.GetUrl()))
		mux.Handle(urlToReplace, vHost)
		redirectToWWW(urlToReplace, mux)
		if setVHostOptions != nil {
			setVHostOptions(vHost)
		}
	}
}

func transformMap(virtualHosts interface{}) map[string]hosts.IVirtualHost {
	result := make(map[string]hosts.IVirtualHost)

	switch t := virtualHosts.(type) {
	case map[string]*hosts.WebVirtualHost:
		for n, v := range t {
			result[n] = v
		}
	case map[string]*hosts.SshVirtualHost:
		for n, v := range t {
			result[n] = v
		}
	case map[string]*hosts.GrpcJsonVirtualHost:
		for n, v := range t {
			result[n] = v
		}
	case map[string]*hosts.GrpcWebVirtualHost:
		for n, v := range t {
			result[n] = v
		}
	default:
		v := reflect.ValueOf(virtualHosts)
		if v.Kind() == reflect.Map {
			for _, key := range v.MapKeys() {
				strct := v.MapIndex(key)
				logs.Log.Info(fmt.Sprint(key.Interface(), strct.Interface()))
			}
		}
	}
	return result
}
