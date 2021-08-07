package hosts

import (
	"github.com/janmbaco/go-infrastructure/errors/errorschecker"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/sshutil"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
	"net/http"
	"strconv"
)

const SSHVirtualHostTenant = "SshVirtualHost"

// SSHVirtualHost is used to configure a virtual host with a web client and a ssh server.
type SSHVirtualHost struct {
	VirtualHost
	KnownHosts string `json:"known_hosts"`
	proxy      sshutil.Proxy
}

func SSHVirtualHostProvider(host *SSHVirtualHost, proxy sshutil.Proxy, logger logs.Logger) IVirtualHost {
	host.proxy = proxy
	host.logger = logger
	return host
}

func (sshVirtualHost *SSHVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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
	hostKeyCallBack, err := knownhosts.New(sshVirtualHost.KnownHosts)
	errorschecker.TryPanic(err)
	clientConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: hostKeyCallBack,
	}
	clientConn, err := ssh.Dial("tcp", sshVirtualHost.HostName+":"+strconv.Itoa(int(sshVirtualHost.Port)), clientConfig)
	errorschecker.TryPanic(err)
	defer func() {
		sshVirtualHost.logger.PrintError(logs.Error, clientConn.Close())
	}()
	sshServerConfig := &ssh.ServerConfig{NoClientAuth: true}
	sshKey, err := ssh.ParsePrivateKey(sshutil.MockSSHKey[:])
	errorschecker.TryPanic(err)
	sshServerConfig.AddHostKey(sshKey)
	_, err = rw.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	sshVirtualHost.logger.PrintError(logs.Error, err)
	conn, _, err := rw.(http.Hijacker).Hijack()
	errorschecker.TryPanic(err)
	sshVirtualHost.proxy.Initialize(conn, sshServerConfig, clientConn)
	sshVirtualHost.proxy.Serve()
}
