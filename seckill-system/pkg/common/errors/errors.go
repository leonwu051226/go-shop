package errors

import "errors"

// 通用错误码
type ErrorCode int

const (
	OK                  ErrorCode = 0
	ErrInternal         ErrorCode = 500
	ErrBadRequest       ErrorCode = 400
	ErrUnauthorized     ErrorCode = 401
	ErrForbidden        ErrorCode = 403
	ErrNotFound         ErrorCode = 404
	ErrConflict         ErrorCode = 409
	ErrTooManyRequests  ErrorCode = 429
	ErrServiceUnavailable ErrorCode = 503
)

// 通用业务错误
var (
	ErrInvalidParam     = errors.New("invalid parameter")
	ErrRecordNotFound   = errors.New("record not found")
	ErrDuplicateRecord  = errors.New("duplicate record")
	ErrDatabase         = errors.New("database error")
	ErrCache            = errors.New("cache error")
	ErrMQ               = errors.New("message queue error")
	ErrTokenInvalid     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrPermissionDenied = errors.New("permission denied")
)
