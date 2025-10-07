package constants

type PaymentStatusString string

const (
	PendingPaymentStatus    PaymentStatusString = "pending"
	SettlementPaymentStatus PaymentStatusString = "settlement"
	ExpiredPaymentStatus    PaymentStatusString = "expired"
)
