package jsrest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrBadRequest                   = NewHTTPError(http.StatusBadRequest)
	ErrUnauthorized                 = NewHTTPError(http.StatusUnauthorized)
	ErrPaymentRequired              = NewHTTPError(http.StatusPaymentRequired)
	ErrForbidden                    = NewHTTPError(http.StatusForbidden)
	ErrNotFound                     = NewHTTPError(http.StatusNotFound)
	ErrMethodNotAllowed             = NewHTTPError(http.StatusMethodNotAllowed)
	ErrNotAcceptable                = NewHTTPError(http.StatusNotAcceptable)
	ErrProxyAuthRequired            = NewHTTPError(http.StatusProxyAuthRequired)
	ErrRequestTimeout               = NewHTTPError(http.StatusRequestTimeout)
	ErrConflict                     = NewHTTPError(http.StatusConflict)
	ErrGone                         = NewHTTPError(http.StatusGone)
	ErrLengthRequired               = NewHTTPError(http.StatusLengthRequired)
	ErrPreconditionFailed           = NewHTTPError(http.StatusPreconditionFailed)
	ErrRequestEntityTooLarge        = NewHTTPError(http.StatusRequestEntityTooLarge)
	ErrRequestURITooLong            = NewHTTPError(http.StatusRequestURITooLong)
	ErrUnsupportedMediaType         = NewHTTPError(http.StatusUnsupportedMediaType)
	ErrRequestedRangeNotSatisfiable = NewHTTPError(http.StatusRequestedRangeNotSatisfiable)
	ErrExpectationFailed            = NewHTTPError(http.StatusExpectationFailed)
	ErrTeapot                       = NewHTTPError(http.StatusTeapot)
	ErrMisdirectedRequest           = NewHTTPError(http.StatusMisdirectedRequest)
	ErrUnprocessableEntity          = NewHTTPError(http.StatusUnprocessableEntity)
	ErrLocked                       = NewHTTPError(http.StatusLocked)
	ErrFailedDependency             = NewHTTPError(http.StatusFailedDependency)
	ErrTooEarly                     = NewHTTPError(http.StatusTooEarly)
	ErrUpgradeRequired              = NewHTTPError(http.StatusUpgradeRequired)
	ErrPreconditionRequired         = NewHTTPError(http.StatusPreconditionRequired)
	ErrTooManyRequests              = NewHTTPError(http.StatusTooManyRequests)
	ErrRequestHeaderFieldsTooLarge  = NewHTTPError(http.StatusRequestHeaderFieldsTooLarge)
	ErrUnavailableForLegalReasons   = NewHTTPError(http.StatusUnavailableForLegalReasons)

	ErrInternalServerError           = NewHTTPError(http.StatusInternalServerError)
	ErrNotImplemented                = NewHTTPError(http.StatusNotImplemented)
	ErrBadGateway                    = NewHTTPError(http.StatusBadGateway)
	ErrServiceUnavailable            = NewHTTPError(http.StatusServiceUnavailable)
	ErrGatewayTimeout                = NewHTTPError(http.StatusGatewayTimeout)
	ErrHTTPVersionNotSupported       = NewHTTPError(http.StatusHTTPVersionNotSupported)
	ErrVariantAlsoNegotiates         = NewHTTPError(http.StatusVariantAlsoNegotiates)
	ErrInsufficientStorage           = NewHTTPError(http.StatusInsufficientStorage)
	ErrLoopDetected                  = NewHTTPError(http.StatusLoopDetected)
	ErrNotExtended                   = NewHTTPError(http.StatusNotExtended)
	ErrNetworkAuthenticationRequired = NewHTTPError(http.StatusNetworkAuthenticationRequired)
)

type HTTPError struct {
	Code    int
	Message string
}

func NewHTTPError(code int) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: http.StatusText(code),
	}
}

func (err *HTTPError) Error() string {
	return fmt.Sprintf("[%d] %s", err.Code, err.Message)
}

func WriteError(w http.ResponseWriter, err error) {
	je := ToJSONError(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(je.Code)

	enc := json.NewEncoder(w)
	_ = enc.Encode(je)
}

func Errorf(he *HTTPError, format string, a ...any) error {
	err := fmt.Errorf(format, a...) //nolint:goerr113
	return errors.Join(he, err)
}

type JSONError struct {
	Code     int      `json:"-"`
	Messages []string `json:"messages"`
}

func ToJSONError(err error) *JSONError {
	je := &JSONError{
		Code: 500,
	}
	je.importError(err)

	return je
}

type singleUnwrap interface {
	Unwrap() error
}

type multiUnwrap interface {
	Unwrap() []error
}

func (je *JSONError) importError(err error) {
	je.Messages = append(je.Messages, err.Error())

	// Pre-traversal for error codes, so we get the most detailed in the stack
	if he, ok := err.(*HTTPError); ok { //nolint:errorlint
		je.Code = he.Code
	}

	if unwrap, ok := err.(singleUnwrap); ok { //nolint:errorlint
		je.importError(unwrap.Unwrap())
	} else if unwrap, ok := err.(multiUnwrap); ok { //nolint:errorlint
		for _, sub := range unwrap.Unwrap() {
			je.importError(sub)
		}
	}
}
