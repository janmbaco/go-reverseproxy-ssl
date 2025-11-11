package hosts_resolver

import (
	"testing"

	"github.com/janmbaco/go-infrastructure/v2/dependencyinjection"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/domain"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/infrastructure/grpcutil"
	"github.com/janmbaco/go-reverseproxy-ssl/v3/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVirtualHostResolver_WhenCalled_ThenReturnsResolver(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}

	// Act
	resolver := NewVirtualHostResolver(container, logger)

	// Assert
	assert.NotNil(t, resolver)
}

func TestVirtualHostResolver_Resolve_WhenValidWebVirtualHosts_ThenReturnsHosts(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}
	resolver := NewVirtualHostResolver(container, logger)

	config := &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "example.com",
						Scheme:   "http",
						HostName: "backend",
						Port:     8080,
					},
				},
			},
		},
	}

	// Act
	hosts, err := resolver.Resolve(config)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, hosts, 1)
	assert.Equal(t, "example.com", hosts[0].GetFrom())
}

func TestVirtualHostResolver_Resolve_WhenDuplicateWebVirtualHost_ThenReturnsError(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}
	resolver := NewVirtualHostResolver(container, logger)

	config := &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "example.com",
						Scheme:   "http",
						HostName: "backend1",
						Port:     8080,
					},
				},
			},
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "example.com", // Duplicate
						Scheme:   "http",
						HostName: "backend2",
						Port:     8081,
					},
				},
			},
		},
	}

	// Act
	hosts, err := resolver.Resolve(config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, hosts)

	vhErr, ok := err.(VirtualHostResolverError)
	require.True(t, ok, "Error should be VirtualHostResolverError")
	assert.Equal(t, VirtualHostDuplicateError, vhErr.GetErrorType())
}

func TestVirtualHostResolver_Resolve_WhenValidGrpcWebVirtualHosts_ThenReturnsHosts(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}
	resolver := NewVirtualHostResolver(container, logger)

	config := &domain.Config{
		GrpcWebVirtualHosts: []*domain.GrpcWebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "grpc.example.com",
						Scheme:   "http",
						HostName: "grpc-backend",
						Port:     50051,
					},
				},
				GrpcWebProxy: &grpcutil.GrpcWebProxy{
					GrpcProxy: grpcutil.GrpcProxy{
						IsTransparentServer: true,
					},
				},
			},
		},
	}

	// Act
	hosts, err := resolver.Resolve(config)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, hosts, 1)
	assert.Equal(t, "grpc.example.com", hosts[0].GetFrom())
}

func TestVirtualHostResolver_Resolve_WhenMixedVirtualHosts_ThenReturnsAllHosts(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}
	resolver := NewVirtualHostResolver(container, logger)

	config := &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "web.example.com",
						Scheme:   "http",
						HostName: "web-backend",
						Port:     8080,
					},
				},
			},
		},
		GrpcWebVirtualHosts: []*domain.GrpcWebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "grpc.example.com",
						Scheme:   "http",
						HostName: "grpc-backend",
						Port:     50051,
					},
				},
				GrpcWebProxy: &grpcutil.GrpcWebProxy{
					GrpcProxy: grpcutil.GrpcProxy{
						IsTransparentServer: true,
					},
				},
			},
		},
	}

	// Act
	hosts, err := resolver.Resolve(config)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, hosts, 2)

	// Verify both hosts are present (order not guaranteed)
	froms := []string{hosts[0].GetFrom(), hosts[1].GetFrom()}
	assert.Contains(t, froms, "web.example.com")
	assert.Contains(t, froms, "grpc.example.com")
}

func TestVirtualHostResolver_Resolve_WhenDuplicateGrpcWebVirtualHost_ThenReturnsError(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}
	resolver := NewVirtualHostResolver(container, logger)

	config := &domain.Config{
		GrpcWebVirtualHosts: []*domain.GrpcWebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "grpc.example.com",
						Scheme:   "http",
						HostName: "grpc-backend1",
						Port:     50051,
					},
				},
				GrpcWebProxy: &grpcutil.GrpcWebProxy{
					GrpcProxy: grpcutil.GrpcProxy{
						IsTransparentServer: true,
					},
				},
			},
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "grpc.example.com", // Duplicate
						Scheme:   "http",
						HostName: "grpc-backend2",
						Port:     50052,
					},
				},
				GrpcWebProxy: &grpcutil.GrpcWebProxy{
					GrpcProxy: grpcutil.GrpcProxy{
						IsTransparentServer: false,
					},
				},
			},
		},
	}

	// Act
	hosts, err := resolver.Resolve(config)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, hosts)

	vhErr, ok := err.(VirtualHostResolverError)
	require.True(t, ok, "Error should be VirtualHostResolverError")
	assert.Equal(t, VirtualHostDuplicateError, vhErr.GetErrorType())
}

func TestVirtualHostResolver_Resolve_WhenEmptyConfig_ThenReturnsEmptyList(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}
	resolver := NewVirtualHostResolver(container, logger)

	config := &domain.Config{
		WebVirtualHosts:     []*domain.WebVirtualHost{},
		GrpcWebVirtualHosts: []*domain.GrpcWebVirtualHost{},
	}

	// Act
	hosts, err := resolver.Resolve(config)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, hosts)
}

func TestVirtualHostResolver_Resolve_WhenMultipleResolves_ThenEachResolveIsIndependent(t *testing.T) {
	// Arrange
	container := dependencyinjection.NewContainer()
	logger := &mocks.MockLogger{}
	resolver := NewVirtualHostResolver(container, logger)

	config1 := &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "config1.com",
						Scheme:   "http",
						HostName: "backend1",
						Port:     8080,
					},
				},
			},
		},
	}

	config2 := &domain.Config{
		WebVirtualHosts: []*domain.WebVirtualHost{
			{
				ClientCertificateHost: domain.ClientCertificateHost{
					VirtualHostBase: domain.VirtualHostBase{
						From:     "config2.com",
						Scheme:   "http",
						HostName: "backend2",
						Port:     8080,
					},
				},
			},
		},
	}

	// Act
	hosts1, err1 := resolver.Resolve(config1)
	hosts2, err2 := resolver.Resolve(config2)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, hosts1, 1)
	assert.Len(t, hosts2, 1)
	assert.Equal(t, "config1.com", hosts1[0].GetFrom())
	assert.Equal(t, "config2.com", hosts2[0].GetFrom())
}
