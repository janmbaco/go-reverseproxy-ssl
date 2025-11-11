package hosts_resolver

import (
	"fmt"
	"reflect"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/grpcutil"
	"google.golang.org/grpc"
)

const (
	_host              = "host"
	_clientCertificate = "clientCertificate"
	_hostName          = "hostName"
	_grpcWebProxy      = "grpcWebProxy"
)

// VirtualHostResolver defines and object responsible to mantain all virtualhost from configuration
type VirtualHostResolver interface {
	Resolve(config *domain.Config) ([]domain.IVirtualHost, error)
}

type virtualHostResolver struct {
	virtualHostsByFrom      map[string]domain.IVirtualHost
	certificateByServerName map[string]domain.CertificateProvider
	resolver                dependencyinjection.Resolver
}

// NewVirtualHostResolver creates a VirtualHostResolver
func NewVirtualHostResolver(container dependencyinjection.Container, logger domain.Logger) VirtualHostResolver {

	container.Register().AsSingleton(new(domain.Logger), func() domain.Logger { return logger }, nil)

	container.Register().AsTenant(domain.WebVirtualHostTenant, new(domain.IVirtualHost), domain.WebVirtualHostProvider, map[int]string{0: _host})
	container.Register().AsTenant(domain.GrpcWebVirtualHostTenant, new(domain.IVirtualHost), domain.GrpcWebVirtualHostProvider, map[int]string{0: _host})

	// register grputil
	container.Register().AsScope(new(*grpc.ClientConn), grpcutil.NewGrpcClientConn, map[int]string{0: _grpcWebProxy, 1: _clientCertificate, 2: _hostName})
	container.Register().AsScope(new(*grpc.Server), grpcutil.NewGrpcServer, map[int]string{0: _grpcWebProxy})
	container.Register().AsScope(new(*grpcweb.WrappedGrpcServer), grpcutil.NewWrappedGrpcServer, map[int]string{0: _grpcWebProxy})

	return &virtualHostResolver{resolver: container.Resolver()}
}

// Resolve resolves the virtual host from the config to the collection
func (vc *virtualHostResolver) Resolve(newConfig *domain.Config) ([]domain.IVirtualHost, error) {

	vc.virtualHostsByFrom = make(map[string]domain.IVirtualHost)
	vc.certificateByServerName = make(map[string]domain.CertificateProvider)

	for _, host := range newConfig.WebVirtualHosts {
		host.EnsureID() // Ensure each virtual host has a unique ID
		h := vc.resolver.Tenant(
			domain.WebVirtualHostTenant,
			new(domain.IVirtualHost),
			map[string]interface{}{_host: host},
		).(domain.IVirtualHost)
		if err := vc.insert(h); err != nil {
			return nil, err
		}
	}

	for _, host := range newConfig.GrpcWebVirtualHosts {
		host.EnsureID() // Ensure each virtual host has a unique ID
		h := vc.resolver.Tenant(
			domain.GrpcWebVirtualHostTenant,
			new(domain.IVirtualHost),
			map[string]interface{}{
				_host:              host,
				_grpcWebProxy:      host.GrpcWebProxy,
				_clientCertificate: host.ClientCertificate,
				_hostName:          host.GetHostName(),
			},
		).(domain.IVirtualHost)
		if err := vc.insert(h); err != nil {
			return nil, err
		}
	}
	result := make([]domain.IVirtualHost, 0)
	for _, vHost := range vc.virtualHostsByFrom {
		result = append(result, vHost)
	}
	return result, nil
}

func (vc *virtualHostResolver) insert(host domain.IVirtualHost) error {
	if _, isContained := vc.virtualHostsByFrom[host.GetFrom()]; isContained {
		return newVirtualHostResolverError(VirtualHostDuplicateError, fmt.Sprintf("the %v virtual host is duplicate in config file", host.GetFrom()), nil)
	}
	vc.virtualHostsByFrom[host.GetFrom()] = host
	if host.GetServerCertificate() != nil {
		if _, isContained := vc.certificateByServerName[host.GetHostToReplace()]; isContained {
			if host.GetServerCertificate() != nil && !reflect.DeepEqual(vc.certificateByServerName[host.GetHostToReplace()], host.GetServerCertificate()) {
				return newVirtualHostResolverError(CertificateDuplicateError, fmt.Sprintf("the %v server name should has always the same certificate", host.GetHostToReplace()), nil)
			}
		} else {
			vc.certificateByServerName[host.GetHostToReplace()] = host.GetServerCertificate()
		}
	}
	return nil
}
