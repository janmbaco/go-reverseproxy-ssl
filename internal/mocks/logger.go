package mocks

import (
	"log"

	"github.com/janmbaco/go-infrastructure/v2/logs"
	"github.com/stretchr/testify/mock"
)

// MockLogger is a mock implementation of Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(message string) {
	m.Called(message)
}

func (m *MockLogger) Error(message string) {
	m.Called(message)
}

func (m *MockLogger) SetDir(dir string) {
	m.Called(dir)
}

func (m *MockLogger) SetConsoleLevel(level logs.LogLevel) {
	m.Called(level)
}

func (m *MockLogger) SetFileLogLevel(level logs.LogLevel) {
	m.Called(level)
}

func (m *MockLogger) GetErrorLogger() *log.Logger {
	args := m.Called()
	return args.Get(0).(*log.Logger)
}

func (m *MockLogger) PrintError(level logs.LogLevel, err error) {
	m.Called(level, err)
}
