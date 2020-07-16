package utils

import (
	"fmt"
)

type ErrorType string

const (
	ErrorReasonSpecUpdate       ErrorType = "SpecUpdate"
	ErrorReasonSpecInvalid      ErrorType = "SpecInvalid"
	ErrorReasonResourceCreate   ErrorType = "ResourceCreate"
	ErrorReasonResourceUpdate   ErrorType = "ResourceUpdate"
	ErrorReasonResourceWaiting  ErrorType = "ResourceWaiting"
	ErrorReasonResourceInvalid  ErrorType = "ResourceInvalid"
	ErrorReasonResourceShutdown ErrorType = "ResourceShutdown"
	ErrorReasonServerWaiting    ErrorType = "ServerWaiting"
	ErrorReasonServerDown       ErrorType = "ServerDown"
	ErrorReasonUnknown          ErrorType = "Unknown"
)

type SQError interface {
	Type() ErrorType
}

type Error struct {
	Reason  ErrorType
	Message string
}

func (r *Error) Type() ErrorType {
	return r.Reason
}

func (r *Error) Error() string {
	return fmt.Sprintf("%s: %s", r.Reason, r.Message)
}

// ReasonForError returns the HTTP status for a particular error.
func ReasonForError(err error) ErrorType {
	switch t := err.(type) {
	case SQError:
		return t.Type()
	}
	return ErrorReasonUnknown
}
