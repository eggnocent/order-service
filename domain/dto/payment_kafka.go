package dto

import (
	"order-service/constants"
	"time"

	"github.com/google/uuid"
)

type PaymentData struct {
	OrderID   uuid.UUID                     `json:"order_id"`
	PaymentID uuid.UUID                     `json:"payment_id"`
	Status    constants.PaymentStatusString `json:"status"`
	ExpiredAt *time.Time                    `json:"expired_at"`
	PaidAt    *time.Time                    `json:"paid_at"`
}

type PaymentContent struct {
	Event    KafkaEvent             `json:"event"`
	Metadata KafkaMetaData          `json:"metadata"`
	Body     KafkaBody[PaymentData] `json:"body"`
}
