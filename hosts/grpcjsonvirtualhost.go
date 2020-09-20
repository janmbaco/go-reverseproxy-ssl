package hosts

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/janmbaco/go-infrastructure/errorhandler"
	"github.com/janmbaco/go-reverseproxy-ssl/grpcUtil"
)

type GrpcJsonVirtualHost struct {
	*ClientCertificateHost
}

func (this *GrpcJsonVirtualHost) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	_, err := url.Parse(this.Scheme + "://" + this.HostName + ":" + strconv.Itoa(int(this.Port)))
	errorhandler.TryPanic(err)
	this.serve(rw, req, func(outReq *http.Request) {
		this.redirectRequest(outReq, req)
	}, grpcUtil.NewTransportJSon(this.ClientCertificate))

}
