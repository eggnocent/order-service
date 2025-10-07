package error

import "errors"

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrFieldAlreadyBooked = errors.New("field schedule already booked")
)

var OrderError = []error{
	ErrOrderNotFound,
	ErrFieldAlreadyBooked,
}
