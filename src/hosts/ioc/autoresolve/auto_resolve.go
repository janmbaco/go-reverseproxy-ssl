package autoresolve

import (
	"github.com/janmbaco/go-infrastructure/dependencyinjection/static"
	// register by inversion of control.
	_ "github.com/janmbaco/go-reverseproxy-ssl/src/hosts/ioc"
	"github.com/janmbaco/go-reverseproxy-ssl/src/hosts/resolver"
)

// GetVirtualHostResolver resolves a VirtualHostResolver
func GetVirtualHostResolver() resolver.VirtualHostResolver {
	return static.Container.Resolver().Type(new(resolver.VirtualHostResolver), nil).(resolver.VirtualHostResolver)
}