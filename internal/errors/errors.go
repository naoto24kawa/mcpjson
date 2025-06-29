package errors

import (
	"fmt"
)

type ErrorType int

const (
	TypeGeneral ErrorType = iota
	TypeValidation
	TypeFile
	TypeNetwork
	TypeConfig
)

type AppError struct {
	Type    ErrorType
	Message string
	Cause   error
	Code    int
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

func NewValidationError(message string) *AppError {
	return &AppError{
		Type:    TypeValidation,
		Message: message,
		Code:    7,
	}
}

func NewFileError(message string, cause error) *AppError {
	return &AppError{
		Type:    TypeFile,
		Message: message,
		Cause:   cause,
		Code:    3,
	}
}

func NewConfigError(message string, cause error) *AppError {
	return &AppError{
		Type:    TypeConfig,
		Message: message,
		Cause:   cause,
		Code:    5,
	}
}

func NewGeneralError(message string, cause error) *AppError {
	return &AppError{
		Type:    TypeGeneral,
		Message: message,
		Cause:   cause,
		Code:    1,
	}
}
