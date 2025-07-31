package types

import "fmt"

// OMIEError represents a custom error type for the OMIE library
type OMIEError struct {
	Code    string
	Message string
	Err     error
}

func (e *OMIEError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *OMIEError) Unwrap() error {
	return e.Err
}

// NewOMIEError creates a new OMIEError
func NewOMIEError(code, message string, err error) *OMIEError {
	return &OMIEError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Common error codes
const (
	ErrCodeDownload     = "DOWNLOAD_ERROR"
	ErrCodeParse        = "PARSE_ERROR"
	ErrCodeInvalidDate  = "INVALID_DATE"
	ErrCodeInvalidData  = "INVALID_DATA"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeNetwork      = "NETWORK_ERROR"
	ErrCodeEncoding     = "ENCODING_ERROR"
)