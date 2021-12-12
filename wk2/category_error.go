package go_camp

import "strings"

const (
	InternalError = "internal error"
	ClientError   = "Client error"

	ErrorNoRows  = "no rows"
	ErrorWrongID = "wrong ID"

	Database = "database"
)

// CategoryError wraps the error from other entities and adds some supplement information.
type CategoryError struct {
	category    string
	source      string // which dependent service contributes this error
	errorCode   string
	subCode     string
	message     string
	internalMsg string
}

func NewInternalError(msg string) *CategoryError {
	return &CategoryError{
		category: InternalError,
		source:   Database,
		message:  msg,
	}
}

func NewClientError(msg string) *CategoryError {
	return &CategoryError{
		category: ClientError,
		source:   Database,
		message:  msg,
	}
}

func (ce *CategoryError) Error() string {
	return ce.message
}

func (ce *CategoryError) Source() string {
	return ce.source
}

func (ce *CategoryError) IsErrorNoRow() bool {
	return strings.EqualFold(ce.message, ErrorNoRows)
}

func (ce *CategoryError) IsClientError() bool {
	return ce.category == ClientError
}

func (ce *CategoryError) IsInternalError() bool {
	return ce.category == InternalError
}
