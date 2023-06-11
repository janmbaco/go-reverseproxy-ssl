package resolver

import (
	"github.com/janmbaco/go-infrastructure/errors"
)

// VirtualHostCollectionError is the struct of an error occurs in FileConfigHandler object
type VirtualHostResolverError interface {
	errors.CustomError
	GetErrorType() VirtualHostCollectionErrorType
}

type virtualHostResolverError struct {
	*errors.CustomizableError
	ErrorType VirtualHostCollectionErrorType
}

func newVirtualHostResolverError(errorType VirtualHostCollectionErrorType, message string, internalError error) VirtualHostResolverError {
	return &virtualHostResolverError{
		ErrorType: errorType,
		CustomizableError: &errors.CustomizableError{
			Message:       message,
			InternalError: internalError,
		}}
}

func (v *virtualHostResolverError) GetErrorType() VirtualHostCollectionErrorType {
	return v.ErrorType
}

// VirtualHostCollectionErrorType is the type of the errors of VirtualHostResolver
type VirtualHostCollectionErrorType uint8

const (
	// UnexpectedError unexpected error
	UnexpectedError VirtualHostCollectionErrorType = iota
	// VirtualHostDuplicateError virtual host duplicated error
	VirtualHostDuplicateError
	// CertificateDuplicateError certificate duplicate error
	CertificateDuplicateError
)