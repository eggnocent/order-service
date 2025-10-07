package dto

import (
	"order-service/constants"
	"time"

	"github.com/google/uuid"
)

type OrderRequest struct {
	FieldScheduleIDs []string `json:"fieldScheduleIDs" validate:"required"`
}

type OrderRequestParam struct {
	Page       int     `json:"page" validate:"required"`
	Limit      int     `json:"limit" validate:"required"`
	SortColumn *string `json:"sortColumn"`
	SortOrder  *string `json:"sortOrder"`
}

type OrderResponse struct {
	UUID        uuid.UUID                   `json:"uuid"`
	Code        string                      `json:"code"`
	UserName    string                      `json:"userName"`
	Amount      float64                     `json:"amount"`
	Status      constants.OrderStatusString `json:"status"`
	PaymentLink string                      `json:"paymentLink"`
	OrderDate   time.Time                   `json:"orderDate"`
	CreatedAt   time.Time                   `json:"createdAt"`
	UpdatedAt   time.Time                   `json:"updatedAt"`
}

type OrderByUserIDResponse struct {
	Code        string                      `json:"code"`
	Amount      string                      `json:"amount"`
	Status      constants.OrderStatusString `json:"status"`
	OrderDate   time.Time                   `json:"orderDate"`
	PaymentLink string                      `json:"paymentLink"`
	InvoiceLink *string                     `json:"invoiceLink"`
}
