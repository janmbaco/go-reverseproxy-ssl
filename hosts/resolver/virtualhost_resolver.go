package resolver

import (
	"fmt"
	"github.com/janmbaco/go-infrastructure/dependencyinjection"
	"github.com/janmbaco/go-infrastructure/errors"
	"github.com/janmbaco/go-reverseproxy-ssl/configs"
	"github.com/janmbaco/go-reverseproxy-ssl/configs/certs"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts"
	"reflect"
)

// VirtualHostResolver defines and object responsible to mantain all virtualhost from configuration
type VirtualHostResolver interface {
	Resolve(newConfig *configs.Config)
	VirtualHosts() []hosts.IVirtualHost
}

type virtualHostResolver struct {
	virtualHosts            []hosts.IVirtualHost
	virtualHostsByFrom      map[string]hosts.IVirtualHost
	certificateByServerName map[string]*certs.CertificateDefs
	deferer                 errors.ErrorDefer
	resolver                dependencyinjection.Resolver
}

func NewVirtualHostCollection(thrower errors.ErrorThrower, resolver dependencyinjection.Resolver) VirtualHostResolver {
	return &virtualHostResolver{deferer: errors.NewErrorDefer(thrower, &virtualHostCollectionErrorPipe{}), resolver: resolver}
}

// Resolve resolves the virtual host from the config to the collection
func (vc *virtualHostResolver) Resolve(newConfig *configs.Config) {
	defer vc.deferer.TryThrowError()

	vc.virtualHosts = make([]hosts.IVirtualHost, 0)
	vc.virtualHostsByFrom = make(map[string]hosts.IVirtualHost)
	vc.certificateByServerName = make(map[string]*certs.CertificateDefs)

	for _, host := range newConfig.WebVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.WebVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{"host": host},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.SshVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.SSHVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{"host": host},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.GrpcJSONVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.GrpcWebVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{
				"host":              host,
				"clientCertificate": host.ClientCertificate,
			},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.GrpcWebVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.GrpcWebVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{
				"host":              host,
				"grpcWebProxy":      &host.GrpcWebProxy,
				"grpcProxy":         &host.GrpcProxy,
				"clientCertificate": host.ClientCertificate,
				"hostName":          host.GetHostName(),
			},
		).(hosts.IVirtualHost))
	}

	for _, host := range newConfig.GrpcVirtualHosts {
		vc.insert(vc.resolver.Tenant(
			hosts.GrpcVirtualHostTenant,
			new(hosts.IVirtualHost),
			map[string]interface{}{
				"host":              host,
				"grpcProxy":         &host.GrpcProxy,
				"clientCertificate": host.ClientCertificate,
				"hostName":          host.GetHostName(),
			},
		).(hosts.IVirtualHost))
	}
}

// VirtualHosts gets all virtual host registered
func (vc *virtualHostResolver) VirtualHosts() []hosts.IVirtualHost {
	return vc.virtualHosts
}

func (vc *virtualHostResolver) insert(host hosts.IVirtualHost) {
	if _, isContained := vc.virtualHostsByFrom[host.GetFrom()]; isContained {
		panic(newVirtualHostCollectionError(VirtualHostDuplicateError, fmt.Sprintf("the %v virtual host is duplicate in config file", host.GetFrom())))
	}
	vc.virtualHostsByFrom[host.GetFrom()] = host
	if host.GetServerCertificate() != nil {
		if _, isContained := vc.certificateByServerName[host.GetHostToReplace()]; isContained {
			if host.GetServerCertificate() != nil && !reflect.DeepEqual(vc.certificateByServerName[host.GetHostToReplace()], host.GetServerCertificate()) {
				panic(newVirtualHostCollectionError(CerteficateDuplicateError, fmt.Sprintf("the %v server name should has always the same certificate", host.GetHostToReplace())))
			}
		} else {
			vc.certificateByServerName[host.GetHostToReplace()] = host.GetServerCertificate()
		}
	}
	vc.virtualHosts = append(vc.virtualHosts, host)
}
