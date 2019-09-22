package main

import (
	"crypto/tls"
	"fmt"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/servers"
	"net/http"
	"strings"
)

func main() {
	//redirect http to https
	go func() {
		servers.NewListener(redirectHttpToHttps).Start()
	}()
	//start server
	servers.NewListener(reverseProxy).Start()
}


func redirectHttpToHttps(httpServer *http.Server){
	cross.Log.Info("Start Redirect Server from http to https")
	mux := http.NewServeMux()
	mux.Handle("/",  http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusMovedPermanently)
	}))
	httpServer.Addr = ":80"
	httpServer.Handler = mux
}

func reverseProxy(httpServer *http.Server){

		cross.Log.SetDir(servers.Config.LogsDir)
		cross.Log.SetConsoleLevel(servers.Config.LogConsoleLevel)
		cross.Log.SetFileLogLevel(servers.Config.LogFileLevel)

		cross.Log.Info("")
		cross.Log.Info("Start Server Application")
		cross.Log.Info("")

		//create a Multiplexer server
		mux := http.NewServeMux()

		var virtualHost []string
		var caPems []string
		var isRegisteredDefaultHost bool

		for name, vHost := range servers.Config.VirtualHost{

			virtualHost = append(virtualHost, name)

			if len(vHost.CaPem) > 0 {
				caPems = append(caPems, vHost.CaPem)
			}

			if name == servers.Config.DefaultHost{
				isRegisteredDefaultHost = true
			}
			cross.Log.Info(fmt.Sprintf("register proxy from: '%v' to '%v://%v:%v'", name,vHost.Scheme, vHost.HostName, vHost.Port))
			mux.Handle(name+"/", vHost)

			if strings.Contains(name, "www") {
				//redirect to web host with www.
				mux.Handle(strings.Replace(name, "www.", "", 1)+"/",
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						http.Redirect(w, r, "https://"+name, http.StatusMovedPermanently)
					}))
			}
		}

		if !isRegisteredDefaultHost {
			cross.Log.Info(fmt.Sprintf("register default host: '%v'", servers.Config.DefaultHost))
			mux.HandleFunc(servers.Config.DefaultHost + "/", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("en curso..."))
			})
			virtualHost = append(virtualHost, servers.Config.DefaultHost)
			if strings.Contains(servers.Config.DefaultHost, "www") {
				//redirect to web host with www.
				mux.Handle(strings.Replace(servers.Config.DefaultHost, "www.", "", 1)+"/",
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						http.Redirect(w, r, "https://"+servers.Config.DefaultHost, http.StatusMovedPermanently)
					}))
			}
		}

		httpServer.Addr = servers.Config.ReverseProxyPort
		httpServer.Handler = mux
		httpServer.TLSConfig = servers.GetTlsConfig(virtualHost, caPems)
		if servers.Config.DefaultHost == "localhost" {
			httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
		}
	}

