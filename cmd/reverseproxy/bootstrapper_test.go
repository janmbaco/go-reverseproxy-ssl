package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewServerBootstrapper_WhenCalled_ThenReturnsBootstrapper(t *testing.T) {
	// Arrange & Act
	bootstrapper := NewServerBootstrapper()

	// Assert
	assert.NotNil(t, bootstrapper)
}

func TestServerBootstrapper_BuildContainer_WhenCalled_ThenReturnsValidContainer(t *testing.T) {
	// Arrange
	bootstrapper := NewServerBootstrapper()

	// Act
	container := bootstrapper.BuildContainer()

	// Assert
	assert.NotNil(t, container)
	assert.NotNil(t, container.Resolver())
	assert.NotNil(t, container.Register())
}

func TestServerBootstrapper_CreateDefaultConfig_WhenCalled_ThenReturnsDefaultConfig(t *testing.T) {
	// Arrange
	bootstrapper := &ServerBootstrapper{}

	// Act
	config := bootstrapper.CreateDefaultConfig()

	// Assert
	assert.NotNil(t, config)
	assert.Equal(t, "localhost", config.DefaultHost)
	assert.Equal(t, ":443", config.ReverseProxyPort)
	assert.Equal(t, ":8081", config.ConfigUIPort)
	assert.Len(t, config.WebVirtualHosts, 1)
	assert.Equal(t, "www.example.com", config.WebVirtualHosts[0].From)
}
