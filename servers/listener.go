package servers

import (
	"context"
	"errors"
	"github.com/janmbaco/go-reverseproxy-ssl/cross"
	"github.com/janmbaco/go-reverseproxy-ssl/events"
	"net/http"
)

type ConfigureListenerFunc func(httpServer *http.Server)

type Listener struct {
	configureFunc ConfigureListenerFunc
	reStart       bool
	onConfigChanged events.SubscribeFunc
	httpServer *http.Server
}

func NewListener(configureFunc ConfigureListenerFunc) *Listener{
	return &Listener{
		configureFunc: configureFunc,
	}
}

func (this *Listener) Start() {
	for this.httpServer == nil || this.reStart {
		this.reStart = false
		if this.onConfigChanged != nil{
			events.UnSubscribe("ConfigFileChanged", &this.onConfigChanged)
		}
		this.httpServer = &http.Server{
			ErrorLog: cross.Log.ErrorLogger,
		}
		if this.configureFunc == nil{
			cross.TryPanic(errors.New("Not configured HttpServer"))
		}
		this.configureFunc(this.httpServer)
		if this.onConfigChanged == nil {
			this.onConfigChanged = func(args *events.EventArgs) {
				this.reStart = true
				this.Stop()
			}
		}
		events.Subscribe("ConfigFileChanged", &this.onConfigChanged)
		cross.Log.Info("Listen on "+this.httpServer.Addr)
		err := this.httpServer.ListenAndServeTLS("", "")
		cross.Log.Error(err.Error())
	}
}


func (this *Listener) Stop() error {
	cross.Log.Info("Server Stop")
	return this.httpServer.Shutdown(context.Background())
}



