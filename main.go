package main

import (
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/disk"
	"github.com/janmbaco/go-reverseproxy-ssl/servers"
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type(
	remoteHost struct{
		Scheme string `json:"scheme"`
		HostName string `json:"host_name"`
		Port uint `json:"port"`
	}

	config struct{
		VirtualHost map[string]remoteHost `json:"virtual_hosts"`
		DefaultHost string `json:"default_host"`
		ReverseProxyPort string `json:"reverse_proxy_port"`
		LogConsoleLevel cross.LogLevel `json:"log_console_level"`
		LogFileLevel    cross.LogLevel `json:"log_file_level"`
	}
)

func main() {
	cross.Log.Info("")
	cross.Log.Info("Start Server Application")
	cross.Log.Info("")

	//default config
	conf := &config{
		VirtualHost:  map[string]remoteHost{
			"example.host.com" : {
				Scheme: "http",
				HostName: "localhost",
				Port: 2256,
			},
		},
		ReverseProxyPort: ":443",
		LogConsoleLevel:  cross.Trace,
		LogFileLevel:     cross.Warning,
	}
	//write o read config from file
	disk.LoadConfig(conf)

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/",  http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
		}))
		server := &http.Server{
			Addr:              ":80",
			Handler:           mux,
			ErrorLog:          cross.Log.ErrorLogger,
		}
		cross.Log.Info("Start Redirect Server from http to https")
		cross.TryPanic(server.ListenAndServe())
	}()

	getProxy := func (scheme string, host string, port uint)  *httputil.ReverseProxy  {
		remote, err := url.Parse(scheme + "://" + host + ":" + strconv.Itoa(int(port)))
		cross.TryPanic(err)
		return httputil.NewSingleHostReverseProxy(remote)
	}

	servers.NewListener(
		func(httpServer *http.Server)  {
			mux := http.NewServeMux()
			//redirect to web host with www.
			mux.Handle(strings.Replace(conf.DefaultHost, "www.","", 1)+"/",
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "https://"+conf.DefaultHost+r.RequestURI, http.StatusMovedPermanently)
				}))
			var virtualHost []string
			for name, vhost := range conf.VirtualHost{
				virtualHost = append(virtualHost, name)
				mux.Handle(name + "/", getProxy(vhost.Scheme, vhost.HostName, vhost.Port))
			}
			autocert := &autocert.Manager{
				Cache:      autocert.DirCache("certs"),
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist(virtualHost...),
			}
			httpServer.Addr = conf.ReverseProxyPort
			httpServer.Handler = mux
			httpServer.TLSConfig = autocert.TLSConfig()
		}).Start()
}

