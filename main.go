package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/janmbaco/go-reverseproxy-ssl/configs"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/disk"
	"github.com/janmbaco/go-reverseproxy-ssl/servers"
)

var Config *configs.Config

func main() {

	setConfiguration()
	//redirect http to https
	go func() {
		servers.NewListener(redirectHttpToHttps).Start()
	}()
	//start server
	servers.NewListener(reverseProxy).Start()
}

func setConfiguration(){
	//default config if file is not found
	Config = &configs.Config{
		VirtualHost:  map[string]*configs.VirtualHost{
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
		LogsDir: "../logs",
	}

	var configfile = flag.String("ConfigFile", "./" + filepath.Base(os.Args[0]) + ".config", "Config File")

	var configConstructor  = func() interface{}{
		return &configs.Config{}
	}

	var configCopy = func(from interface{}, to interface{}) {
		fromConf := from.(*configs.Config)
		toConf := to.(*configs.Config)
		*toConf = *fromConf
	}

	disk.NewConfigFileManager(*configfile, configConstructor, configCopy).Load(Config)
}

func redirectHttpToHttps(serverSetter *servers.ServerSetter) {
	cross.Log.Info("Start Redirect Server from http to https")
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	}))
	serverSetter.Addr = ":80"
	serverSetter.Handler = mux
}

func reverseProxy(serverSetter *servers.ServerSetter) {

	cross.Log.SetDir(Config.LogsDir)
	cross.Log.SetConsoleLevel(Config.LogConsoleLevel)
	cross.Log.SetFileLogLevel(Config.LogFileLevel)

	cross.Log.Info("")
	cross.Log.Info("Start Server Application")
	cross.Log.Info("")

	//create a Multiplexer server
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

		cross.Log.Info(fmt.Sprintf("register proxy from: '%v' to '%v://%v:%v'", name, vHost.Scheme, vHost.HostName, vHost.Port))
		mux.Handle(name+"/", vHost)
		redirectToWWW(name, mux)
	}

	if !isRegisteredDefaultHost {
		cross.Log.Info(fmt.Sprintf("register default host: '%v'", Config.DefaultHost))
		mux.HandleFunc(Config.DefaultHost+"/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("started..."))
		})
		virtualHost = append(virtualHost, Config.DefaultHost)

		if Config.DefaultHost != "localhost"{
			redirectToWWW(Config.DefaultHost, mux)
		}
	}

	serverSetter.Addr = Config.ReverseProxyPort
	serverSetter.Handler = mux
	serverSetter.TLSConfig = configs.GetTlsConfig(Config, virtualHost, caPems)
	if Config.DefaultHost == "localhost" {
		serverSetter.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
	}
}

func redirectToWWW(hostname string, mux *http.ServeMux){
	if strings.Contains(hostname, "www") {
		//redirect to web host with www.
		mux.Handle(strings.Replace(hostname, "www.", "", 1)+"/",
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://"+hostname, http.StatusMovedPermanently)
			}))
	}
}
