package jsrest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Code int

const (
	StatusBadRequest                   Code = http.StatusBadRequest
	StatusUnauthorized                 Code = http.StatusUnauthorized
	StatusPaymentRequired              Code = http.StatusPaymentRequired
	StatusForbidden                    Code = http.StatusForbidden
	StatusNotFound                     Code = http.StatusNotFound
	StatusMethodNotAllowed             Code = http.StatusMethodNotAllowed
	StatusNotAcceptable                Code = http.StatusNotAcceptable
	StatusProxyAuthRequired            Code = http.StatusProxyAuthRequired
	StatusRequestTimeout               Code = http.StatusRequestTimeout
	StatusConflict                     Code = http.StatusConflict
	StatusGone                         Code = http.StatusGone
	StatusLengthRequired               Code = http.StatusLengthRequired
	StatusPreconditionFailed           Code = http.StatusPreconditionFailed
	StatusRequestEntityTooLarge        Code = http.StatusRequestEntityTooLarge
	StatusRequestURITooLong            Code = http.StatusRequestURITooLong
	StatusUnsupportedMediaType         Code = http.StatusUnsupportedMediaType
	StatusRequestedRangeNotSatisfiable Code = http.StatusRequestedRangeNotSatisfiable
	StatusExpectationFailed            Code = http.StatusExpectationFailed
	StatusTeapot                       Code = http.StatusTeapot
	StatusMisdirectedRequest           Code = http.StatusMisdirectedRequest
	StatusUnprocessableEntity          Code = http.StatusUnprocessableEntity
	StatusLocked                       Code = http.StatusLocked
	StatusFailedDependency             Code = http.StatusFailedDependency
	StatusTooEarly                     Code = http.StatusTooEarly
	StatusUpgradeRequired              Code = http.StatusUpgradeRequired
	StatusPreconditionRequired         Code = http.StatusPreconditionRequired
	StatusTooManyRequests              Code = http.StatusTooManyRequests
	StatusRequestHeaderFieldsTooLarge  Code = http.StatusRequestHeaderFieldsTooLarge
	StatusUnavailableForLegalReasons   Code = http.StatusUnavailableForLegalReasons

	StatusInternalServerError           Code = http.StatusInternalServerError
	StatusNotImplemented                Code = http.StatusNotImplemented
	StatusBadGateway                    Code = http.StatusBadGateway
	StatusServiceUnavailable            Code = http.StatusServiceUnavailable
	StatusGatewayTimeout                Code = http.StatusGatewayTimeout
	StatusHTTPVersionNotSupported       Code = http.StatusHTTPVersionNotSupported
	StatusVariantAlsoNegotiates         Code = http.StatusVariantAlsoNegotiates
	StatusInsufficientStorage           Code = http.StatusInsufficientStorage
	StatusLoopDetected                  Code = http.StatusLoopDetected
	StatusNotExtended                   Code = http.StatusNotExtended
	StatusNetworkAuthenticationRequired Code = http.StatusNetworkAuthenticationRequired
)

type Error struct {
	Code     Code
	Messages []string
	Params   map[string]any
}

func NewError() *Error {
	return &Error{
		Params: map[string]any{},
	}
}

func FromError(err error, code Code) *Error {
	e := NewError()
	e.Code = code

	e.importError(err)

	return e
}

func Errorf(code Code, format string, a ...any) *Error {
	err := fmt.Errorf(format, a...) //nolint:goerr113
	return FromError(err, code)
}

func (e *Error) SetParam(key string, value any) {
	e.Params[key] = value
}

func (e *Error) Error() string {
	msg, err := json.Marshal(e.Values())
	if err != nil {
		return err.Error()
	}

	return string(msg)
}

func (e *Error) Write(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(e.Code))

	enc := json.NewEncoder(w)
	enc.Encode(e.Values()) //nolint:errcheck,errchkjson
}

func (e *Error) Values() map[string]any {
	vals := map[string]any{
		"errors": e.Messages,
	}

	for k, v := range e.Params {
		vals[k] = v
	}

	return vals
}

type singleUnwrap interface {
	Unwrap() error
}

type multiUnwrap interface {
	Unwrap() []error
}

func (e *Error) importError(err error) {
	e.Messages = append(e.Messages, err.Error())

	if unwrap, ok := err.(singleUnwrap); ok { //nolint:errorlint
		e.importError(unwrap.Unwrap())
	} else if unwrap, ok := err.(multiUnwrap); ok { //nolint:errorlint
		for _, sub := range unwrap.Unwrap() {
			e.importError(sub)
		}
	}
}
