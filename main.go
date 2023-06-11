package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-infrastructure/server"
	"github.com/janmbaco/go-reverseproxy-ssl/src/configs"
	"github.com/janmbaco/go-reverseproxy-ssl/src/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/src/hosts"
	"github.com/janmbaco/go-reverseproxy-ssl/src/hosts/ioc/autoresolve"
	"golang.org/x/crypto/acme/autocert"

	fileConfigResolver "github.com/janmbaco/go-infrastructure/configuration/fileconfig/ioc/resolver"
	logsResolver "github.com/janmbaco/go-infrastructure/logs/ioc/resolver"
	serverResolver "github.com/janmbaco/go-infrastructure/server/ioc/resolver"
)

func main() {
	var configFile = flag.String("config", "", "globalConfig File")
	flag.Parse()
	if len(*configFile) == 0 {
		_, _ = fmt.Fprint(os.Stderr, "You must set a config file!\n")
		flag.PrintDefaults()
		return
	}

	listenerBuilder := serverResolver.GetListenerBuilder(
		fileConfigResolver.GetFileConfigHandler(
			*configFile,
			&configs.Config{
				WebVirtualHosts: []*hosts.WebVirtualHost{
					{
						ClientCertificateHost: hosts.ClientCertificateHost{
							VirtualHost: hosts.VirtualHost{
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
			},
		),
	)

	// start server
	finish := listenerBuilder.SetBootstrapper(reverseProxy).GetListener().Start()

	// redirect http to https
	listenerBuilder.SetBootstrapper(redirectHTTPToHTTPS).GetListener().Start()

	errorschecker.TryPanic(<-finish)
}

func redirectHTTPToHTTPS(config interface{}, serverSetter *server.ServerSetter) {
	logger := logsResolver.GetLogger()
	if serverSetter.IsChecking {
		logger.SetConsoleLevel(logs.Warning)
		logger.SetFileLogLevel(logs.Warning)
	} else {
		setLogConfiguration(config.(*configs.Config), logger)
	}
	logger.Info("")
	logger.Info("Start Redirect Server from http to https")
	logger.Info("")
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	}))
	serverSetter.Addr = ":80"
	serverSetter.Handler = mux
}

func reverseProxy(config interface{}, serverSetter *server.ServerSetter) {
	logger := logsResolver.GetLogger()
	if serverSetter.IsChecking {
		logger.SetConsoleLevel(logs.Warning)
		logger.SetFileLogLevel(logs.Warning)
	} else {
		setLogConfiguration(config.(*configs.Config), logger)
	}
	vhCollection := autoresolve.GetVirtualHostResolver().Resolve(config.(*configs.Config))

	logger.Info("")
	logger.Info("Start Server Application")
	logger.Info("")

	mux := http.NewServeMux()

	certManager := certs.NewCertManager(&autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./certs"),
	})

	registerVirtualHost(mux, certManager, vhCollection, logger)

	logger.Info(fmt.Sprintf("register default host: '%v'", config.(*configs.Config).DefaultHost))
	mux.HandleFunc(config.(*configs.Config).DefaultHost+"/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("started..."))
	})

	if config.(*configs.Config).DefaultHost != "localhost" {
		redirectToWWW(config.(*configs.Config).DefaultHost, mux)
	}
	serverSetter.Addr = config.(*configs.Config).ReverseProxyPort
	serverSetter.Handler = mux
	serverSetter.TLSConfig = certManager.GetTLSConfig()
}
func setLogConfiguration(config *configs.Config, logger logs.Logger) {
	logger.SetDir(config.LogsDir)
	logger.SetConsoleLevel(config.LogConsoleLevel)
	logger.SetFileLogLevel(config.LogFileLevel)
}
func registerVirtualHost(mux *http.ServeMux, certManager *certs.CertManager, virtualHosts []hosts.IVirtualHost, logs logs.Logger) {
	for _, vHost := range virtualHosts {
		vHost.SetURLToReplace()
		urlToReplace := vHost.GetURLToReplace()
		logs.Info(fmt.Sprintf("register proxy from: '%v' to %v", vHost.GetFrom(), vHost.GetURL()))
		mux.Handle(urlToReplace, vHost)
		if isRegisterdedCertificate := certManager.HasCertificateFor(vHost.GetHostToReplace()); !isRegisterdedCertificate && vHost.GetServerCertificate() == nil {
			certManager.AddAutoCertificate(vHost.GetFrom())
		} else if !isRegisterdedCertificate {
			certManager.AddCertificate(vHost.GetHostToReplace(), vHost.GetServerCertificate().GetCertificate())
		}
		certManager.AddClientCA(vHost.GetAuthorizedCAs())
		redirectToWWW(urlToReplace, mux)
	}
}

func redirectToWWW(hostname string, mux *http.ServeMux) {
	if strings.HasPrefix(hostname, "www") {
		mux.Handle(strings.Replace(hostname, "www.", "", 1),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+hostname, http.StatusMovedPermanently)
			}))
	}
}
