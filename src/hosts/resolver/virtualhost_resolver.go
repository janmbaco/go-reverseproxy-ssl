package resolver

import (
	"fmt"
	"reflect"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/janmbaco/go-infrastructure/dependencyinjection"
	"github.com/janmbaco/go-infrastructure/errors"
	"github.com/janmbaco/go-reverseproxy-ssl/src/configs"
	"github.com/janmbaco/go-reverseproxy-ssl/src/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/src/grpcutil"
	"github.com/janmbaco/go-reverseproxy-ssl/src/hosts"
	"github.com/janmbaco/go-reverseproxy-ssl/src/sshutil"
	"google.golang.org/grpc"
)

const (
	_host              = "host"
	_clientCertificate = "clientCertificate"
	_grpcProxy         = "grpcProxy"
	_hostName          = "hostName"
	_grpcWebProxy      = "grpcWebProxy"
)

// VirtualHostResolver defines and object responsible to mantain all virtualhost from configuration
type VirtualHostResolver interface {
	Resolve(newConfig *configs.Config) []hosts.IVirtualHost
}

type virtualHostResolver struct {
	virtualHostsByFrom      map[string]hosts.IVirtualHost
	certificateByServerName map[string]*certs.CertificateDefs
	deferer                 errors.ErrorDefer
	resolver                dependencyinjection.Resolver
}

// NewVirtualHostResolver creates a VirtualHostResolver
func NewVirtualHostResolver(errorDefer errors.ErrorDefer, container dependencyinjection.Container) VirtualHostResolver {

	container.Register().AsTenant(hosts.WebVirtualHostTenant, new(hosts.IVirtualHost), hosts.WebVirtualHostProvider, map[uint]string{0: _host})
	container.Register().AsTenant(hosts.SSHVirtualHostTenant, new(hosts.IVirtualHost), hosts.SSHVirtualHostProvider, map[uint]string{0: _host})
	container.Register().AsTenant(hosts.GrpcJSONVirtualHostTenant, new(hosts.IVirtualHost), hosts.GrpcJSONVirtualHostProvider, map[uint]string{0: _host})
	container.Register().AsTenant(hosts.GrpcVirtualHostTenant, new(hosts.IVirtualHost), hosts.GrpcVirtualHostProvider, map[uint]string{0: _host})
	container.Register().AsTenant(hosts.GrpcWebVirtualHostTenant, new(hosts.IVirtualHost), hosts.GrpcWebVirtualHostProvider, map[uint]string{0: _host})

	// register sshutil
	container.Register().AsScope(new(sshutil.Proxy), sshutil.NewProxy, nil)

	// register grputil
	container.Register().AsScope(new(grpcutil.TransportJSON), grpcutil.NewTransportJSON, map[uint]string{0: _clientCertificate})
	container.Register().AsScope(new(*grpc.ClientConn), grpcutil.NewGrpcClientConn, map[uint]string{0: _grpcProxy, 1: _clientCertificate, 2: _hostName})
	container.Register().AsScope(new(*grpc.Server), grpcutil.NewGrpcServer, map[uint]string{0: _grpcProxy})
	container.Register().AsScope(new(*grpcweb.WrappedGrpcServer), grpcutil.NewWrappedGrpcServer, map[uint]string{0: _grpcWebProxy})

	return &virtualHostResolver{deferer: errorDefer, resolver: container.Resolver()}
}

// Resolve resolves the virtual host from the config to the collection
func (vc *virtualHostResolver) Resolve(newConfig *configs.Config) []hosts.IVirtualHost {
	defer vc.deferer.TryThrowError(vc.pipeError)

	vc.virtualHostsByFrom = make(map[string]hosts.IVirtualHost)
	vc.certificateByServerName = make(map[string]*certs.CertificateDefs)

	for _, host := range newConfig.WebVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.WebVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{_host: host},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.SSHVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.SSHVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{_host: host},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.GrpcJSONVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.GrpcWebVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{
				_host:              host,
				_clientCertificate: host.ClientCertificate,
			},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.GrpcWebVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.GrpcWebVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{
				_host:              host,
				_grpcWebProxy:      &host.GrpcWebProxy,
				_grpcProxy:         &host.GrpcProxy,
				_clientCertificate: host.ClientCertificate,
				_hostName:          host.GetHostName(),
			},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.GrpcVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.GrpcVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{
				_host:              host,
				_grpcProxy:         &host.GrpcProxy,
				_clientCertificate: host.ClientCertificate,
				_hostName:          host.GetHostName(),
			},
		).(hosts.IVirtualHost))
	}
	result := make([]hosts.IVirtualHost, 0)
	for _, vHost := range vc.virtualHostsByFrom {
		result = append(result, vHost)
	}
	return result
}

func (vc *virtualHostResolver) insert(host hosts.IVirtualHost) {
	if _, isContained := vc.virtualHostsByFrom[host.GetFrom()]; isContained {
		panic(newVirtualHostResolverError(VirtualHostDuplicateError, fmt.Sprintf("the %v virtual host is duplicate in config file", host.GetFrom()), nil))
	}
	vc.virtualHostsByFrom[host.GetFrom()] = host
	if host.GetServerCertificate() != nil {
		if _, isContained := vc.certificateByServerName[host.GetHostToReplace()]; isContained {
			if host.GetServerCertificate() != nil && !reflect.DeepEqual(vc.certificateByServerName[host.GetHostToReplace()], host.GetServerCertificate()) {
				panic(newVirtualHostResolverError(CertificateDuplicateError, fmt.Sprintf("the %v server name should has always the same certificate", host.GetHostToReplace()), nil))
			}
		} else {
			vc.certificateByServerName[host.GetHostToReplace()] = host.GetServerCertificate()
		}
	}
}

func (vc *virtualHostResolver) pipeError(err error) error {
	resultError := err

	if errType := reflect.Indirect(reflect.ValueOf(err)).Type(); !errType.Implements(reflect.TypeOf((*VirtualHostResolverError)(nil)).Elem()) {
		resultError = newVirtualHostResolverError(UnexpectedError, err.Error(), err)
	}

	return resultError
}