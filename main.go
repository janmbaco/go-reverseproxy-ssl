package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-infrastructure/server"
	"github.com/janmbaco/go-reverseproxy-ssl/configs"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
	"golang.org/x/crypto/acme/autocert"
)

var globalConfig *configs.Config //nolint:gochecknoglobals
var virtualHosts []hosts.IVirtualHost

func main() {
	var configFile = flag.String("config", "", "globalConfig File")
	flag.Parse()
	if len(*configFile) == 0 {
		_, _ = fmt.Fprint(os.Stderr, "You must set a config file!\n")
		flag.PrintDefaults()
		return
	}

	globalConfig = configs.NewConfig(
		configs.NewConfigHandler(*configFile),
		"localhost",
		":443",
		logs.Trace,
		logs.Trace,
		"./logs")

	recollectVirtualHostByConfig(globalConfig)
	globalConfig.OnModifyingConfigSubscriber(recollectVirtualHostByConfig)

	// redirect http to https
	go func() {
		server.NewListener(globalConfig, redirectHttpToHttps).Start()
	}()
	// start server
	server.NewListener(globalConfig, reverseProxy).Start()
}

func setLogConfiguration() {
	logs.Log.SetDir(globalConfig.LogsDir)
	logs.Log.SetConsoleLevel(globalConfig.LogConsoleLevel)
	logs.Log.SetFileLogLevel(globalConfig.LogFileLevel)
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

	setLogConfiguration()

	logs.Log.Info("")
	logs.Log.Info("Start Server Application")
	logs.Log.Info("")

	mux := http.NewServeMux()

	certManager := certs.NewCertManager(&autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./certs"),
	})

	registerVirtualHost(mux, certManager, virtualHosts)

	logs.Log.Info(fmt.Sprintf("register default host: '%v'", globalConfig.DefaultHost))
	mux.HandleFunc(globalConfig.DefaultHost+"/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("started..."))
	})

	if globalConfig.DefaultHost != "localhost" {
		redirectToWWW(globalConfig.DefaultHost, mux)
	}

	serverSetter.Addr = globalConfig.ReverseProxyPort
	serverSetter.Handler = mux
	serverSetter.TLSConfig = certManager.GetTlsConfig()

}

func redirectToWWW(hostname string, mux *http.ServeMux) {

	if strings.HasPrefix(hostname, "www") {
		// redirect to web host with www.
		mux.Handle(strings.Replace(hostname, "www.", "", 1),
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+hostname, http.StatusMovedPermanently)
			}))
	}

}

func registerVirtualHost(mux *http.ServeMux, certManager *certs.CertManager, virtualHosts []hosts.IVirtualHost) {

	for _, vHost := range virtualHosts {
		vHost.SetUrlToReplace()
		urlToReplace := vHost.GetUrlToReplace()
		logs.Log.Info(fmt.Sprintf("register proxy from: '%v' to %v", vHost.GetFrom(), vHost.GetUrl()))
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

func recollectVirtualHostByConfig(newConfig interface{}) {
	virtualHostByFrom := make(map[string]hosts.IVirtualHost, 0)
	certificateByServerName := make(map[string]certs.CertificateDefs, 0)
	virtualHosts = make([]hosts.IVirtualHost, 0)
	verifyAndInser := func(host hosts.IVirtualHost) {
		if _, isContained := virtualHostByFrom[host.GetFrom()]; isContained {
			panic(fmt.Sprintf("The %v virtual host is duplicate in config file!!", host.GetFrom()))
		}
		virtualHostByFrom[host.GetFrom()] = host
		if _, isContained := certificateByServerName[host.GetHostToReplace()]; isContained {
			if host.GetServerCertificate() != nil && certificateByServerName[host.GetHostToReplace()] != *host.GetServerCertificate() {
				panic(fmt.Sprintf("The %v server name should has always the same certificate!!", host.GetHostToReplace()))
			}
		} else {
			certificateByServerName[host.GetHostToReplace()] = *host.GetServerCertificate()
		}
		virtualHosts = append(virtualHosts, host)
	}
	for _, v := range newConfig.(*configs.Config).WebVirtualHosts {
		verifyAndInser(v)
	}

	for _, v := range newConfig.(*configs.Config).SshVirtualHosts {
		verifyAndInser(v)
	}

	for _, v := range newConfig.(*configs.Config).GrpcJsonVirtualHosts {
		verifyAndInser(v)
	}

	for _, v := range newConfig.(*configs.Config).GrpcWebVirtualHosts {
		verifyAndInser(v)
	}

	for _, v := range newConfig.(*configs.Config).GrpcVirtualHosts {
		verifyAndInser(v)
	}
}
