package servers

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"

	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/events"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type (
	ConfigureListenerFunc func(serverSetter *ServerSetter)

 	SetProtobufFunc func(grpcServer *grpc.Server)

 	ServerType uint8

 	ServerSetter struct {
		Addr      string
		Handler      http.Handler
		TLSConfig    *tls.Config
		TLSNextProto map[string]func(*http.Server, *tls.Conn, http.Handler)
		ServerType   ServerType	
	}

 	Listener struct {
		configureFunc   ConfigureListenerFunc
		setProtobufFunc SetProtobufFunc
		reStart         bool
		onConfigChanged events.SubscribeFunc
		serverSetter    *ServerSetter
		httpServer      *http.Server
		grpcServer      *grpc.Server
	}
)

const (
	HttpServer ServerType = iota
	gRpcSever
)

func NewListener(configureFunc ConfigureListenerFunc) *Listener {
	return &Listener{
		configureFunc: configureFunc,
		serverSetter: &ServerSetter{},
	}
}

func (this *Listener) SetProtobuf(setProtbufFunc SetProtobufFunc) *Listener {
	this.setProtobufFunc = setProtbufFunc
	this.serverSetter.ServerType = gRpcSever
	return this
}

func (this *Listener) Start() {
	for (this.httpServer == nil && this.grpcServer == nil) || this.reStart {

		this.reStart = false

		if this.configureFunc == nil {
			cross.TryPanic(errors.New("Not configured Server"))
		}
		this.configureFunc(this.serverSetter)

		this.initializeServer()

		this.registerEvents()

		cross.Log.Info("Listen on " + this.serverSetter.Addr)
		var err error
		switch this.serverSetter.ServerType {
		case HttpServer:
			if this.serverSetter.TLSConfig != nil {
				err = this.httpServer.ListenAndServeTLS("", "")
			} else {
				err = this.httpServer.ListenAndServe()
			}
		case gRpcSever:
			var lis net.Listener
			lis, err = net.Listen("tcp", this.serverSetter.Addr)
			cross.TryPanic(err)
			if(this.setProtobufFunc != nil){
				this.setProtobufFunc(this.grpcServer)
			}
			err = this.grpcServer.Serve(lis)
		}
		cross.Log.Warning(err.Error())
	}
}

func (this *Listener) Stop() error {
	cross.Log.Info("Server Stop")
	var err error
	switch this.serverSetter.ServerType {
	case HttpServer:
		err =  this.httpServer.Shutdown(context.Background())
	case gRpcSever:
		this.grpcServer.GracefulStop()
	}
	return err
}

func (this *Listener) initializeServer() {
	if this.serverSetter.Addr == ""{
		cross.TryPanic(errors.New("Address not configured."))
	}
	switch this.serverSetter.ServerType {
	case HttpServer:
		if this.serverSetter.Handler == nil{
			cross.TryPanic(errors.New("Handler routes not configured."))
		}
		this.httpServer = &http.Server{
			ErrorLog: cross.Log.ErrorLogger,
		}
		this.httpServer.Addr = this.serverSetter.Addr
		this.httpServer.Handler = this.serverSetter.Handler
		this.httpServer.TLSConfig = this.serverSetter.TLSConfig
		if this.serverSetter.TLSNextProto != nil{
			this.httpServer.TLSNextProto = this.serverSetter.TLSNextProto
		}

	case gRpcSever:
		opts := []grpc.ServerOption{grpc.Creds(credentials.NewTLS(this.serverSetter.TLSConfig))}
		this.grpcServer = grpc.NewServer(opts...)
	}
}

func(this *Listener) registerEvents(){
	if this.onConfigChanged == nil {
		this.onConfigChanged = func(args *events.EventArgs) {
			this.reStart = true
			this.Stop()
		}
	}

	events.Subscribe("ConfigFileChanged", &this.onConfigChanged)
}