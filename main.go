package main

import (
	"flag"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/janmbaco/go-infrastructure/config"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-infrastructure/server"
	"github.com/janmbaco/go-reverseproxy-ssl/configs"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
	"golang.org/x/crypto/acme/autocert"
)

var Config *configs.Config

func main() {

	var configfile = flag.String("config", "go_reverseproxy_ssl.config", "Config File")
	flag.Parse()

	Config = setDefaultConfig()
	configHandler := config.NewFileConfigHandler(*configfile)
	configHandler.Load(Config)
	logConfiguration := setLogConfiguration
	logConfiguration()
	configHandler.OnModifiedConfigSubscriber(&logConfiguration)

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
		WebVirtualHosts: map[string]*hosts.WebVirtualHost{
			"www.example.com": {
				ClientCertificateHost: hosts.ClientCertificateHost{
					VirtualHost: hosts.VirtualHost{
						Scheme:   "http",
						HostName: "localhost",
						Port:     8080,
						ServerCertificate: &certs.CertificateDefs{
							CaPem:      "./certs/CA-cert.pem",
							PublicKey:  "./certs/www.example.com.crt",
							PrivateKey: "./certs/www.example.com.key",
						},
					},
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

	certManager := certs.NewCertManager(&autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./certs"),
	})

	registerVirtualHost(mux, certManager, transformMap(Config.WebVirtualHosts))
	registerVirtualHost(mux, certManager, transformMap(Config.GrpcVirtualHosts))
	registerVirtualHost(mux, certManager, transformMap(Config.GrpcJsonVirtualHosts))
	registerVirtualHost(mux, certManager, transformMap(Config.GrpcWebVirtualHosts))
	registerVirtualHost(mux, certManager, transformMap(Config.SshVirtualHosts))

	logs.Log.Info(fmt.Sprintf("register default host: '%v'", Config.DefaultHost))
	mux.HandleFunc(Config.DefaultHost+"/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("started..."))
	})

	if Config.DefaultHost != "localhost" {
		redirectToWWW(Config.DefaultHost, mux)
	}

	serverSetter.Addr = Config.ReverseProxyPort
	serverSetter.Handler = mux
	serverSetter.TLSConfig = certManager.GetTlsConfig()

}

func redirectToWWW(hostname string, mux *http.ServeMux) {
	if strings.HasPrefix(hostname, "www") {
		//redirect to web host with www.
		mux.Handle(strings.Replace(hostname, "www.", "", 1),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+hostname, http.StatusMovedPermanently)
			}))
	}
}

func registerVirtualHost(mux *http.ServeMux, certManager *certs.CertManager, virtualHosts map[string]hosts.IVirtualHost) {
	for name, vHost := range virtualHosts {
		vHost.SetUrlToReplace(name)
		urlToReplace := vHost.GetUrlToReplace()
		logs.Log.Info(fmt.Sprintf("register proxy from: '%v' to %v", name, vHost.GetUrl()))
		mux.Handle(urlToReplace, vHost)
		if vHost.IsAutoCert() {
			certManager.AddAutoCertificate(vHost.GetHostToReplace())
		} else {
			certManager.AddCertificate(vHost.GetHostToReplace(), vHost.GetServerCertificate())
		}
		certManager.AddClientCA(vHost.GetAuthorizedCAs())
		redirectToWWW(urlToReplace, mux)
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
	case map[string]*hosts.GrpcVirtualHost:
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
