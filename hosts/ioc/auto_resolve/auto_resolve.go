package auto_resolve

import (
	_ "github.com/janmbaco/go-reverseproxy-ssl/hosts/ioc"
	"github.com/janmbaco/go-infrastructure/dependencyinjection/static"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts/resolver"
)

func GetVirtualHostResolver() resolver.VirtualHostResolver {
	return static.Container.Resolver().Type(new(resolver.VirtualHostResolver), nil).(resolver.VirtualHostResolver)
}