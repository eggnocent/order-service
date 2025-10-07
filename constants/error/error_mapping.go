package error

import (
	errPayment "order-service/constants/error/order"
)

func ErrMapping(err error) bool {
	var (
		GeneralErrors = GeneralErrrors
		PaymentErrors = errPayment.OrderError
	)

	allErrors := make([]error, 0)
	allErrors = append(allErrors, GeneralErrors...)
	allErrors = append(allErrors, PaymentErrors...)

	for _, item := range allErrors {
		if err.Error() == item.Error() {
			return true
		}
	}

	return false
}
