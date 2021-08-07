package resolver

import (
	"github.com/janmbaco/go-infrastructure/errors"
	"reflect"
)

// VirtualHostCollectionError is the struct of an error occurs in FileConfigHandler object
type VirtualHostCollectionError struct {
	errors.CustomError
	ErrorType VirtualHostCollectionErrorType
}

func newVirtualHostCollectionError(errorType VirtualHostCollectionErrorType, message string) *VirtualHostCollectionError {
	return &VirtualHostCollectionError{
		ErrorType: errorType,
		CustomError: errors.CustomError{
			Message:       message,
			InternalError: nil,
		}}
}

// VirtualHostCollectionErrorType is the type of the errors of VirtualHostResolver
type VirtualHostCollectionErrorType uint8

const (
	UnexpectedError VirtualHostCollectionErrorType = iota
	VirtualHostDuplicateError
	CerteficateDuplicateError
)

type virtualHostCollectionErrorPipe struct{}

func (*virtualHostCollectionErrorPipe) Pipe(err error) error {
	resultError := err

	if errType := reflect.Indirect(reflect.ValueOf(err)).Type(); errType != reflect.TypeOf(&VirtualHostCollectionError{}) {
		errorType := UnexpectedError
		resultError = &VirtualHostCollectionError{
			CustomError: errors.CustomError{
				Message:       err.Error(),
				InternalError: err,
			},
			ErrorType: errorType,
		}
	}
	return resultError
}
