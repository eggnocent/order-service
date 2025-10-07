package error

import "errors"

var (
	ErrInternalServerError = errors.New("internal server error")
	ErrSQLError            = errors.New("database server failed to execute the query")
	ErrToManyRequests      = errors.New("too many requests")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidToken        = errors.New("invalid token")
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidUploadFile   = errors.New("invalid upload file")
	ErrSizeTooLarge        = errors.New("size too large")
)

var GeneralErrrors = []error{
	ErrInternalServerError,
	ErrSQLError,
	ErrToManyRequests,
	ErrUnauthorized,
	ErrInvalidToken,
	ErrForbidden,
	ErrInvalidUploadFile,
	ErrSizeTooLarge,
}
