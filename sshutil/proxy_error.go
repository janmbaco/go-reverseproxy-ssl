package sshutil

import (
	"github.com/janmbaco/go-infrastructure/errors"
	"reflect"
)

// ProxyError is the struct of an error occurs in FileConfigHandler object
type ProxyError struct {
	errors.CustomError
	ErrorType ProxyErrorType
}

func newProxyError(errorType ProxyErrorType, message string) *ProxyError {
	return &ProxyError{
		ErrorType: errorType,
		CustomError: errors.CustomError{
			Message:       message,
			InternalError: nil,
		}}
}

// ProxyErrorType is the type of the errors of Proxy
type ProxyErrorType uint8

const (
	UnexpectedError ProxyErrorType = iota
	NotInitializedError
)

type proxyErrorPipe struct{}

func (*proxyErrorPipe) Pipe(err error) error {
	resultError := err

	if errType := reflect.Indirect(reflect.ValueOf(err)).Type(); errType != reflect.TypeOf(&ProxyError{}) {
		errorType := UnexpectedError
		resultError = &ProxyError{
			CustomError: errors.CustomError{
				Message:       err.Error(),
				InternalError: err,
			},
			ErrorType: errorType,
		}
	}
	return resultError
}
