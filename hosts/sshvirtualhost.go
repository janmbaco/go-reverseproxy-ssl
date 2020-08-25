package hosts

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"net/http"
	"strconv"

	"github.com/janmbaco/go-infrastructure/errorhandler"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/sshUtil"
)

type SshVirtualHost struct {
	*VirtualHost
	KnownHosts string `json:"known_hosts"`
}

func (this *SshVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

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
	hostKeyCallBack, err := knownhosts.New(this.KnownHosts)
	errorhandler.TryPanic(err)
	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
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
}
