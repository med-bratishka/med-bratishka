package service

import "fmt"

const (
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeConflict     = 409
	CodeInternal     = 500
)

type ServiceError struct {
	Code        int
	PublicCode  string
	PublicMsg   string
	InternalErr error
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}
	if e.InternalErr != nil {
		return e.InternalErr.Error()
	}
	return e.PublicMsg
}

func (e *ServiceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.InternalErr
}

func newServiceError(code int, err error, publicCode, publicMsg string) error {
	return &ServiceError{
		Code:        code,
		PublicCode:  publicCode,
		PublicMsg:   publicMsg,
		InternalErr: err,
	}
}

func wrapInternal(op string, err error) error {
	return newServiceError(CodeInternal, fmt.Errorf("%s: %w", op, err), "INTERNAL_ERROR", "internal server error")
}
