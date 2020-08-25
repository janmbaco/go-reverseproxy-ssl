package hosts

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net/http"
	"net/url"
	"strconv"

	"github.com/janmbaco/go-infrastructure/errorhandler"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/grpcUtil"
	"github.com/mwitkow/grpc-proxy/proxy"
)

type GrpcWebVirtualHost struct {
	*VirtualHost
	*grpcUtil.GrpcWebProxy
	TlsDefs   *certs.TlsDefs `json:"tls_config"`
	Authority string         `json:"authority"`
}

func (this *GrpcWebVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	urlParsed, err := url.Parse(this.Scheme + "://" + this.HostName + ":" + strconv.Itoa(int(this.Port)))
	errorhandler.TryPanic(err)

	var opt []grpc.DialOption
	opt = append(opt, grpc.WithCodec(proxy.Codec()))

	if len(this.Authority) > 0 {
		opt = append(opt, grpc.WithAuthority(this.Authority))
	}

	if this.TlsDefs != nil {
		opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(this.TlsDefs.GetTlsConfig())))
	} else {
		opt = append(opt, grpc.WithInsecure())
	}

	clientConn, err := grpc.Dial(urlParsed.Host, opt...)
	errorhandler.TryPanic(err)

	this.WrappedGrpcServer(clientConn).ServeHTTP(rw, req)

}
