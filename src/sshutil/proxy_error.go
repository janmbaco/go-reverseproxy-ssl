package sshutil

import (
	"github.com/janmbaco/go-infrastructure/errors"
)

// ProxyError is the struct of an error occurs in FileConfigHandler object
type ProxyError interface{
	errors.CustomError
	GetErrorType() ProxyErrorType
}

type proxyError struct {
	*errors.CustomizableError
	ErrorType ProxyErrorType
}

func newProxyError(errorType ProxyErrorType, message string, err error) ProxyError {
	return &proxyError{
		ErrorType: errorType,
		CustomizableError: &errors.CustomizableError{
			Message:       message,
			InternalError: err,
		}}
}

func(p *proxyError) GetErrorType() ProxyErrorType{
	return p.ErrorType
}

// ProxyErrorType is the type of the errors of Proxy
type ProxyErrorType uint8

const (
	UnexpectedError ProxyErrorType = iota
	NotInitializedError
)