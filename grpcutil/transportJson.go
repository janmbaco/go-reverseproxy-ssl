package grpcutil

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"github.com/janmbaco/go-infrastructure/errorhandler"
	"github.com/janmbaco/go-infrastructure/logs"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"golang.org/x/net/http2"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const (
	defaultClientTimeout = time.Second * 60
	headerGRPCStatusCode = "Grpc-Status"
	headerGRPCMessage    = "Grpc-Message"
	headerContentLength  = "Content-Length"
	contentTypeGRPCJSON  = "application/grpc+json"

	grpcNoCompression byte = 0x00
)

// TransportJson is used to transport the communication between a grpc server and a web client (json).
type TransportJson struct {
	TlsDefs *certs.CertificateDefs
}

// NewTransportJson returns a TransportJson
func NewTransportJson(tlsDefs *certs.CertificateDefs) *TransportJson {
	return &TransportJson{TlsDefs: tlsDefs}
}

// RoundTrip return the response in json format.
func (transportJson *TransportJson) RoundTrip(req *http.Request) (*http.Response, error) {

	req = modifyRequestToJSONgRPC(req)
	resp, err := transportJson.getClient().Do(req)
	if err != nil {
		logs.Log.Error(fmt.Sprintf("unable to do request err=[%s]", err))

		buff := bytes.NewBuffer(nil)
		buff.WriteString(err.Error())
		resp = &http.Response{
			StatusCode: 502,
			Body:       ioutil.NopCloser(buff),
		}
		err = nil

	} else {
		resp, err = handleGRPCResponse(resp)
	}
	return resp, err
}

func (transportJson *TransportJson) getClient() *http.Client {
	var client *http.Client
	if transportJson.TlsDefs == nil {
		client = &http.Client{
			// Skip TLS dial
			Transport: &http2.Transport{
				AllowHTTP: true,
				DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(netw, addr)
				},
			},
			Timeout: defaultClientTimeout,
		}
	} else {
		client = &http.Client{
			Transport: &http2.Transport{},
			Timeout:   defaultClientTimeout,
		}
	}
	return client
}

func handleGRPCResponse(resp *http.Response) (*http.Response, error) {

	code := resp.Header.Get(headerGRPCStatusCode)
	if code != "0" && code != "" {
		buff := bytes.NewBuffer(nil)
		grpcMessage := resp.Header.Get(headerGRPCMessage)
		buff.WriteString(fmt.Sprintf(`{"error": %v ,"code": %v}`, grpcMessage, code))

		resp.Body = ioutil.NopCloser(buff)
		resp.StatusCode = 500

		return resp, nil
	}

	prefix := make([]byte, 5)
	_, _ = resp.Body.Read(prefix)

	resp.Header.Del(headerContentLength)

	return resp, nil

}

func modifyRequestToJSONgRPC(req *http.Request) *http.Request {
	// https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md

	var body []byte
	var err error
	// read body so we can add the grpc prefix
	if req.Body != nil {
		body, err = ioutil.ReadAll(req.Body)
		errorhandler.TryPanic(err)
	}
	lenBody := len(body)
	if lenBody < 0 || lenBody > ^(0)-5 {
		panic("invalid request body")
	}
	b := make([]byte, 0, lenBody+5)
	buff := bytes.NewBuffer(b)

	// grpc prefix is
	// 1 byte: compression indicator
	// 4 bytes: content length (excluding prefix)
	errorhandler.TryPanic(buff.WriteByte(grpcNoCompression)) // 0 or 1, indicates compressed payload

	lenBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lenBytes, uint32(len(body)))

	_, err = buff.Write(lenBytes)
	errorhandler.TryPanic(err)
	_, err = buff.Write(body)
	errorhandler.TryPanic(err)

	// create new request
	outReq, err := http.NewRequest(req.Method, req.URL.String(), buff)
	errorhandler.TryPanic(err)
	outReq.Header = req.Header

	// remove content length header
	outReq.Header.Del(headerContentLength)
	outReq.Header.Set("content-type", contentTypeGRPCJSON)

	outReq.RequestURI = ""

	return outReq

}