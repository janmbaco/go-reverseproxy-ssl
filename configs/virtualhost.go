package configs

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"

	"github.com/janmbaco/go-infrastructure/errorhandler"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/sshUtil"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type VirtualHost struct {
	Scheme           string `json:"scheme"`
	HostName         string `json:"host_name"`
	Port             uint   `json:"port"`
	CaPem            string `json:"ca_pem"`
	ClientCrt        string `json:"client_crt"`
	ClientKey        string `json:"client_key"`
	NeedPkFromClient bool   `json:"need_pk_from_client"`
}

func (this *VirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if this.NeedPkFromClient && req.TLS.PeerCertificates == nil {
		http.Error(rw, "Not authorized", 401)
		return
	}
	if strings.ToUpper(this.Scheme) == "SSH" {
		if req.Method != http.MethodConnect {
			http.NotFound(rw, req)
			return
		}
		req.Header.Set("Authorization", req.Header.Get("Proxy-Authorization"))
		user, pass, ok := req.BasicAuth()
		if !ok {
			http.Error(rw, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		hostKeyCallBack, err := knownhosts.New(this.CaPem)
		errorhandler.TryPanic(err)
		clientConfig := &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.Password(string(pass)),
			},
			HostKeyCallback: hostKeyCallBack,
		}
		clientConn, err := ssh.Dial("tcp", this.HostName+":"+strconv.Itoa(int(this.Port)), clientConfig)
		errorhandler.TryPanic(err)
		defer func() {
			logs.Log.TryError(clientConn.Close())
		}()
		sshServerConfig := &ssh.ServerConfig{NoClientAuth: true}
		sshKey, err := ssh.ParsePrivateKey(sshUtil.MockSshKey[:])
		errorhandler.TryPanic(err)
		sshServerConfig.AddHostKey(sshKey)
		_, err = rw.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		logs.Log.TryError(err)
		conn, _, err := rw.(http.Hijacker).Hijack()
		errorhandler.TryPanic(err)
		proxy := sshUtil.Proxy{
			Conn:   conn,
			Config: sshServerConfig,
			Client: clientConn,
		}
		proxy.Serve()

	} else {
		_, err := url.Parse(this.Scheme + "://" + this.HostName + ":" + strconv.Itoa(int(this.Port)))
		errorhandler.TryPanic(err)
		proxy := httputil.ReverseProxy{
			Director: func(outReq *http.Request) {
				outReq.URL.Scheme = this.Scheme
				outReq.URL.Host = this.HostName + ":" + strconv.Itoa(int(this.Port))
				outReq.URL.Path = req.URL.Path
				outReq.URL.RawQuery = req.URL.RawQuery

				outReq.Header.Set("X-Forwarded-Proto", "https")

				logs.Log.Info(fmt.Sprintf("from '%v%v%v' to '%v%v%v'", req.URL.Host, req.URL.Path, req.URL.RawQuery, outReq.URL.Host, outReq.URL.Path, outReq.URL.RawPath))
				if this.NeedPkFromClient {
					pubKey := base64.URLEncoding.EncodeToString(req.TLS.PeerCertificates[0].RawSubjectPublicKeyInfo)
					outReq.Header.Set("X-Forwarded-ClientKey", pubKey)
				}
			},
			Transport:      nil,
			FlushInterval:  0,
			ErrorLog:       logs.Log.ErrorLogger,
			BufferPool:     nil,
			ModifyResponse: nil,
			ErrorHandler:   nil,
		}
		//Add transport tls layer
		transport := http.DefaultTransport.(*http.Transport)
		if len(this.CaPem) > 0 {
			transport.TLSClientConfig = &tls.Config{
				RootCAs: GetCertPool(this.CaPem),
			}
		}
		//add client certificates
		if len(this.ClientKey) > 0 && len(this.ClientCrt) > 0 {
			clientCert, err := tls.LoadX509KeyPair(this.ClientCrt, this.ClientKey)
			errorhandler.TryPanic(err)
			transport.TLSClientConfig.Certificates = []tls.Certificate{clientCert}
		}

		proxy.Transport = transport
		proxy.ServeHTTP(rw, req)
	}

}
