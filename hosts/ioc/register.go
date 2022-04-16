package ioc

import (
	"github.com/janmbaco/go-infrastructure/dependencyinjection/static"
	"github.com/janmbaco/go-reverseproxy-ssl/hosts/resolver"
)

func init() {
	static.Container.Register().AsSingleton(new(resolver.VirtualHostResolver), resolver.NewVirtualHostResolver, nil)
}