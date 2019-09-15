package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/disk"
	"github.com/janmbaco/go-reverseproxy-ssl/servers"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type(
	remoteHost struct{
		Scheme           string `json:"scheme"`
		HostName         string `json:"host_name"`
		Port             uint   `json:"port"`
		CaPem            string `json:"ca_pem"`
		ClientCrt        string `json:"client_crt"`
		ClientKey        string `json:"client_key"`
		NeedPkFromClient bool   `json:"need_pk_from_client"`
	}

	config struct{
		VirtualHost      map[string]*remoteHost `json:"virtual_hosts"`
		DefaultHost      string                `json:"default_host"`
		ReverseProxyPort string                `json:"reverse_proxy_port"`
		LogConsoleLevel  cross.LogLevel        `json:"log_console_level"`
		LogFileLevel     cross.LogLevel        `json:"log_file_level"`
		LogsDir          string                `json:"logs_dir"`
	}
)

func (remoteHost *remoteHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if remoteHost.NeedPkFromClient && req.TLS.PeerCertificates == nil{
		http.Error(rw, "Not authorized", 401)
		return
	}
	_, err := url.Parse(remoteHost.Scheme + "://" + remoteHost.HostName + ":" + strconv.Itoa(int(remoteHost.Port)))
	cross.TryPanic(err)
	proxy := httputil.ReverseProxy{
		Director:        func(outReq *http.Request) {

			outReq.URL.Scheme = remoteHost.Scheme
			outReq.URL.Host = remoteHost.HostName + ":" + strconv.Itoa(int(remoteHost.Port))
			outReq.URL.Path = req.URL.Path
			outReq.URL.RawQuery = req.URL.RawQuery

			outReq.Header.Set("X-Forwarded-Proto", "https")

			cross.Log.Info(fmt.Sprintf( "from '%v%v%v' to '%v%v%v'", req.URL.Host , req.URL.Path, req.URL.RawQuery, outReq.URL.Host, outReq.URL.Path, outReq.URL.RawPath))
			if remoteHost.NeedPkFromClient {
				pubKey := base64.URLEncoding.EncodeToString(req.TLS.PeerCertificates[0].RawSubjectPublicKeyInfo)
				outReq.Header.Set("X-Forwarded-ClientKey", pubKey)
				cross.Log.Info("Public Key usuario: " + pubKey)
			}
		},
		Transport:      nil,
		FlushInterval:  0,
		ErrorLog:       cross.Log.ErrorLogger,
		BufferPool:     nil,
		ModifyResponse: nil,
		ErrorHandler:   nil,
	}
	//Add transport tls layer
	transport := http.DefaultTransport.(*http.Transport)
	if len(remoteHost.CaPem) > 0 {
		transport.TLSClientConfig = &tls.Config{
			RootCAs: getCertPool(remoteHost.CaPem),
		}
	}
	//add client certificates
	if  len(remoteHost.ClientKey) > 0 && len(remoteHost.ClientCrt) > 0 {
		clientCert, err := tls.LoadX509KeyPair(remoteHost.ClientCrt, remoteHost.ClientKey)
		cross.TryPanic(err)
		transport.TLSClientConfig.Certificates=  []tls.Certificate{clientCert}
	}

	proxy.Transport = transport
	proxy.ServeHTTP(rw, req)
}

func  getCertPool (caPems ...string) *x509.CertPool{
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	for _, caPem := range caPems{
		pem, err := ioutil.ReadFile(caPem)
		cross.TryPanic(err)
		rootCAs.AppendCertsFromPEM(pem)
	}
	return rootCAs
}

func main() {
	cross.Log.Info("")
	cross.Log.Info("Start Server Application")
	cross.Log.Info("")

	//default config
	conf := &config{
		VirtualHost:  map[string]*remoteHost{
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
	//añadimos directorio del config
	disk.ConfigFile.SetDir("../configs")
	// añadimos el constructroContenido
	disk.ConfigFile.ConstructorContent = func() interface{}{
			return &config{}
	}

	//añadimo la copia de config
	disk.ConfigFile.CopyContent = func(from interface{}, to interface{}) {
		fromConf := from.(*config)
		toConf := to.(*config)
		*toConf = *fromConf
	}
	//write o read config from file
	disk.ConfigFile.Load(conf)

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

	servers.NewListener(
		func(httpServer *http.Server)  {
			cross.Log.SetDir(conf.LogsDir)
			cross.Log.SetConsoleLevel(conf.LogConsoleLevel)
			cross.Log.SetFileLogLevel(conf.LogFileLevel)


			mux := http.NewServeMux()

			if conf.DefaultHost != "localhost" {
				//redirect to web host with www.
				mux.Handle(strings.Replace(conf.DefaultHost, "www.", "", 1)+"/",
					http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						http.Redirect(w, r, "https://"+conf.DefaultHost+r.RequestURI, http.StatusMovedPermanently)
					}))
			}
			var virtualHost []string
			var caPems []string
			var isRegisteredDefaultHost bool
			for name, vHost := range conf.VirtualHost{
				virtualHost = append(virtualHost, name)
				if len(vHost.CaPem) > 0 {
					caPems = append(caPems, vHost.CaPem)
				}
				if name == conf.DefaultHost{
					isRegisteredDefaultHost = true
				}
				cross.Log.Info(fmt.Sprintf("register proxy from: '%v' to '%v://%v:%v'", name,vHost.Scheme, vHost.HostName, vHost.Port))
				mux.Handle(name+"/", vHost)
			}
			if !isRegisteredDefaultHost {
				cross.Log.Info(fmt.Sprintf("register default host: '%v'", conf.DefaultHost))
				mux.HandleFunc(conf.DefaultHost + "/", func(w http.ResponseWriter, r *http.Request) {
					_, _ = w.Write([]byte("en curso..."))
				})
				virtualHost = append(virtualHost, conf.DefaultHost)
			}

			getTlsConfig := func () *tls.Config{
				ret := &tls.Config{}
				if conf.DefaultHost == "localhost" {
					//in localhost doesn't works autocert
					cert, err := tls.LoadX509KeyPair("../certs/server.crt", "../certs/server.key")
					cross.TryPanic(err)

					ret = &tls.Config{
						Rand:                  rand.Reader,
						Time:                  nil,
						Certificates:          []tls.Certificate{cert},
						NameToCertificate:     nil,
						GetCertificate:        nil,
						GetClientCertificate:  nil,
						GetConfigForClient:    nil,
						VerifyPeerCertificate: nil,
						RootCAs:               nil,
						NextProtos:            nil,
						ServerName:            "",
						ClientCAs:             nil,
						InsecureSkipVerify:    false,
						CipherSuites: []uint16{
							tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
							tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
							tls.TLS_RSA_WITH_AES_256_CBC_SHA,
						},
						PreferServerCipherSuites:    true,
						SessionTicketsDisabled:      false,
						SessionTicketKey:            [32]byte{},
						ClientSessionCache:          nil,
						MinVersion:                  tls.VersionTLS12,
						MaxVersion:                  0,
						CurvePreferences:            []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
						DynamicRecordSizingDisabled: false,
						Renegotiation:               0,
						KeyLogWriter:                nil,
					}
				} else {
					autocert := &autocert.Manager{
						Prompt:          autocert.AcceptTOS,
						Cache:           autocert.DirCache("../certs"),
						HostPolicy:      autocert.HostWhitelist(virtualHost...),
					}

					ret = autocert.TLSConfig()
				}
				ret.ClientAuth=tls.VerifyClientCertIfGiven
				ret.ClientCAs = getCertPool(caPems...)
				return ret
			}
			httpServer.Addr = conf.ReverseProxyPort
			httpServer.Handler = mux
			httpServer.TLSConfig = getTlsConfig()
			if conf.DefaultHost == "localhost" {
				httpServer.TLSNextProto = make(map[string]func(*http.Server, *tls.Conn, http.Handler), 0)
			}
		}).Start()
}
