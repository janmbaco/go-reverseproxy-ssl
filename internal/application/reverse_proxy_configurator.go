package application

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/janmbaco/go-infrastructure/v2/configuration"
	"github.com/janmbaco/go-infrastructure/v2/server"
	"github.com/janmbaco/go-reverseproxy-ssl/internal/domain"
	certs "github.com/janmbaco/go-reverseproxy-ssl/internal/infrastructure/certificates"
	"golang.org/x/crypto/acme/autocert"
)

type ReverseProxyConfigurator struct {
	logger        domain.Logger
	vhResolver    domain.VirtualHostResolver
	configHandler configuration.ConfigHandler
	serverState   *ServerState
}

func NewReverseProxyConfigurator(logger domain.Logger, vhResolver domain.VirtualHostResolver, configHandler configuration.ConfigHandler, serverState *ServerState) *ReverseProxyConfigurator {
	return &ReverseProxyConfigurator{
		logger:        logger,
		vhResolver:    vhResolver,
		configHandler: configHandler,
		serverState:   serverState,
	}
}

func (rpc *ReverseProxyConfigurator) Configure(config interface{}, serverSetter *server.ServerSetter) error {
	cfg := config.(*domain.Config)

	mux := rpc.setupMux()
	certMgr := rpc.setupCertManager(cfg)

	rpc.registerVirtualHosts(mux, certMgr, cfg)

	serverSetter.Addr = cfg.ReverseProxyPort
	serverSetter.Handler = mux
	serverSetter.TLSConfig = certMgr.GetTLSConfig()

	rpc.serverState.UpdateMux(mux)
	rpc.serverState.UpdateCertMgr(certMgr)
	rpc.serverState.UpdateConfig(cfg)

	rpc.logger.Info("")
	rpc.logger.Info("Start Server Application")
	rpc.logger.Info("")

	return nil
}

func (rpc *ReverseProxyConfigurator) setupMux() *http.ServeMux {
	return http.NewServeMux()
}

func (rpc *ReverseProxyConfigurator) setupCertManager(cfg *domain.Config) domain.CertificateManager {
	certMgr := certs.NewCertManager(&autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("./certs"),
	})

	if cfg.DefaultServerCert != "" && cfg.DefaultServerKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.DefaultServerCert, cfg.DefaultServerKey)
		if err != nil {
			rpc.logger.Error(fmt.Sprintf("Failed to load default certificate: %v", err))
		} else {
			certMgr.AddCertificate(cfg.DefaultHost, &cert)
		}
	} else if cfg.DefaultHost != "" {
		certMgr.AddAutoCertificate(cfg.DefaultHost)
	}

	return certMgr
}

func (rpc *ReverseProxyConfigurator) registerVirtualHosts(mux *http.ServeMux, certMgr domain.CertificateManager, cfg *domain.Config) {
	vhCollection, err := rpc.vhResolver.Resolve(cfg)
	if err != nil {
		rpc.logger.Error(fmt.Sprintf("Failed to resolve virtual hosts: %v", err))
		return
	}

	for _, vh := range vhCollection {
		vh.SetURLToReplace()
		urlToReplace := vh.GetURLToReplace()
		rpc.logger.Info(fmt.Sprintf("register proxy from: '%v' to %v", vh.GetFrom(), vh.GetURL()))
		mux.Handle(urlToReplace, vh)

		if !certMgr.HasCertificateFor(vh.GetHostToReplace()) {
			if vh.GetServerCertificate() != nil {
				cert, err := vh.GetServerCertificate().GetCertificate()
				if err != nil {
					rpc.logger.Error(fmt.Sprintf("Failed to get certificate for %v: %v", vh.GetHostToReplace(), err))
					continue
				}
				certMgr.AddCertificate(vh.GetHostToReplace(), cert)
			} else {
				certMgr.AddAutoCertificate(vh.GetFrom())
			}
		}

		certMgr.AddClientCA(vh.GetAuthorizedCAs())
		RedirectToWWW(urlToReplace, mux)
	}

	rpc.registerDefaultHost(mux, vhCollection, cfg)
}

func (rpc *ReverseProxyConfigurator) registerDefaultHost(mux *http.ServeMux, vhCollection []domain.IVirtualHost, cfg *domain.Config) {
	defaultHost := cfg.DefaultHost
	for _, vh := range vhCollection {
		if vh.GetFrom() == defaultHost {
			rpc.logger.Info(fmt.Sprintf("default host '%v' already registered as virtual host, skipping duplicate registration", defaultHost))
			return
		}
	}

	rpc.logger.Info(fmt.Sprintf("register default host: '%v'", defaultHost))
	// ConfigUI is now always registered in the main Configure method
	// rpc.setupConfigUI(mux) // Removed duplicate registration

	if defaultHost != "localhost" {
		RedirectToWWW(defaultHost, mux)
	}
}
