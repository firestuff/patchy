package jsrest

import (
	"encoding/json"
	"errors"
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
}

func FromError(err error, code Code) *Error {
	e := &Error{
		Code: code,
	}

	for iter := err; iter != nil; iter = errors.Unwrap(iter) {
		e.Messages = append(e.Messages, iter.Error())
	}

	return e
}

func (e *Error) Write(w http.ResponseWriter) {
	w.WriteHeader(int(e.Code))

	enc := json.NewEncoder(w)
	enc.Encode(map[string]any{ //nolint:errcheck,errchkjson
		"errors": e.Messages,
	})
}
