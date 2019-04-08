package exception

import (
	"fmt"
	"time"
)

func NewGenericError(err error, c ErrorCode) Error {
	if le, ok := err.(Error); ok {
		return le
	}
	genericError := &GenericError{
		Timestamp: time.Now(),
		Cause:     err,
		ErrorCode: c,
	}
	if err != nil {
		genericError.Message = fmt.Sprintf("[ErrorCode: %s]%s", c.String(), err.Error())
	}
	return genericError
}

func NewGenericErrorWithContext(err error, c ErrorCode, context string) Error {
	if le, ok := err.(Error); ok {
		return le
	}
	genericError := &GenericError{
		Timestamp: time.Now(),
		Cause:     err,
		ErrorCode: c,
	}
	if err != nil {
		genericError.Message = fmt.Sprintf("[ErrorCode: %s, CONTEXT: %s] %s", c.String(), context, err.Error())
	}
	return genericError
}

// Error is the API util type.
type Error interface {
	error
	// Returns the util code for this util.
	Code() ErrorCode
}

type GenericError struct {
	Timestamp time.Time
	ErrorCode ErrorCode
	Cause     error `json:"-"`
	Message   string
}

func (e *GenericError) Error() string {
	return e.Message
}

func (e *GenericError) Code() ErrorCode {
	return e.ErrorCode
}
