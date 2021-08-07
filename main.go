package main

import (
	"flag"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/janmbaco/go-infrastructure/configuration"
	"github.com/janmbaco/go-infrastructure/configuration/fileconfig"
	"github.com/janmbaco/go-infrastructure/dependencyinjection"
	"github.com/janmbaco/go-infrastructure/errors"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-infrastructure/server"
	"github.com/janmbaco/go-reverseproxy-ssl/configs"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/grpcutil"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts/resolver"
	"github.com/janmbaco/go-reverseproxy-ssl/sshutil"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strings"
)

func main() {
	var configFile = flag.String("config", "", "globalConfig File")
	flag.Parse()
	if len(*configFile) == 0 {
		_, _ = fmt.Fprint(os.Stderr, "You must set a config file!\n")
		flag.PrintDefaults()
		return
	}

	container := dependencyinjection.NewContainer()
	registerFacade(container.Register())

	listenerBuilder := container.Resolver().Type(
		new(server.ListenerBuilder),
		map[string]interface{}{"filePath": *configFile, "defaults": &configs.Config{
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
		}},
	).(server.ListenerBuilder)

	logger := container.Resolver().Type(new(logs.Logger), nil).(logs.Logger)
	vhCollection := container.Resolver().Type(new(resolver.VirtualHostResolver), nil).(resolver.VirtualHostResolver)

	// start server
	finish := listenerBuilder.SetBootstrapper(reverseProxy(logger, vhCollection)).GetListener().Start()

	// redirect http to https
	listenerBuilder.SetBootstrapper(redirectHTTPToHTTPS(logger)).GetListener().Start()

	<-finish
}

func registerFacade(register dependencyinjection.Register) {
	register.AsSingleton(new(logs.Logger), logs.NewLogger, nil)
	register.Bind(new(logs.ErrorLogger), new(logs.Logger))
	register.AsSingleton(new(errors.ErrorCatcher), errors.NewErrorCatcher, nil)
	register.AsSingleton(new(errors.ErrorManager), errors.NewErrorManager, nil)
	register.Bind(new(errors.ErrorCallbacks), new(errors.ErrorManager))
	register.AsSingleton(new(errors.ErrorThrower), errors.NewErrorThrower, nil)
	register.AsSingleton(new(configuration.ConfigHandler), fileconfig.NewFileConfigHandler, map[uint]string{0: "filePath", 1: "defaults"})
	register.AsSingleton(new(server.ListenerBuilder), server.NewListenerBuilder, nil)

	// register VitualHosts
	register.AsSingleton(new(resolver.VirtualHostResolver), resolver.NewVirtualHostCollection, nil)
	register.AsTenant(hosts.WebVirtualHostTenant, new(hosts.IVirtualHost), hosts.WebVirtualHostProvider, map[uint]string{0: "host"})
	register.AsTenant(hosts.SSHVirtualHostTenant, new(hosts.IVirtualHost), hosts.SSHVirtualHostProvider, map[uint]string{0: "host"})
	register.AsTenant(hosts.GrpcJSONVirtualHostTenant, new(hosts.IVirtualHost), hosts.GrpcJSONVirtualHostProvider, map[uint]string{0: "host"})
	register.AsTenant(hosts.GrpcVirtualHostTenant, new(hosts.IVirtualHost), hosts.GrpcVirtualHostProvider, map[uint]string{0: "host"})
	register.AsTenant(hosts.GrpcWebVirtualHostTenant, new(hosts.IVirtualHost), hosts.GrpcWebVirtualHostProvider, map[uint]string{0: "host"})

	// register sshutil
	register.AsType(new(sshutil.Proxy), sshutil.NewProxy, nil)

	// register grputil
	register.AsType(new(grpcutil.TransportJSON), grpcutil.NewTransportJSON, map[uint]string{0: "clientCertificate"})
	register.AsType(new(*grpc.ClientConn), grpcutil.NewGrpcClientConn, map[uint]string{0: "grpcProxy", 1: "clientCertificate", 2: "hostName"})
	register.AsType(new(*grpc.Server), grpcutil.NewGrpcServer, map[uint]string{0: "grpcProxy"})
	register.AsType(new(*grpcweb.WrappedGrpcServer), grpcutil.NewWrappedGrpcServer, map[uint]string{0: "grpcWebProxy"})
}

func redirectHTTPToHTTPS(logger logs.Logger) func(interface{}, *server.ServerSetter) {
	return func(config interface{}, serverSetter *server.ServerSetter) {
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
}

func reverseProxy(logger logs.Logger, vhCollection resolver.VirtualHostResolver) func(config interface{}, serverSetter *server.ServerSetter) {
	return func(config interface{}, serverSetter *server.ServerSetter) {
		if serverSetter.IsChecking {
			logger.SetConsoleLevel(logs.Warning)
			logger.SetFileLogLevel(logs.Warning)
		} else {
			setLogConfiguration(config.(*configs.Config), logger)
		}
		vhCollection.Resolve(config.(*configs.Config))

		logger.Info("")
		logger.Info("Start Server Application")
		logger.Info("")

		mux := http.NewServeMux()

		certManager := certs.NewCertManager(&autocert.Manager{
			Prompt: autocert.AcceptTOS,
			Cache:  autocert.DirCache("./certs"),
		})

		registerVirtualHost(mux, certManager, vhCollection.VirtualHosts(), logger)

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
