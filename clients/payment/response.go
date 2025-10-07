package clients

import "github.com/google/uuid"

type PaymentResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}

type PaymentData struct {
	UUID          uuid.UUID `json:"uuid"`
	OrderID       string    `json:"orderID"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	PaymentLink   string    `json:"paymentLink"`
	InvoiceLink   *string   `json:"invoice_link,omitempty"`
	Description   *string   `json:"description"`
	VANumber      *string   `json:"vaNumber,omitempty"`
	Bank          *string   `json:"bank,omitempty"`
	TransactionID *string   `json:"transaction_id,omitempty"`
	Acquirer      *string   `json:"acquirer,omitempty"`
	PaidAt        *string   `json:"paidAt,omitempty"`
	ExpiredAt     *string   `json:"expiredAt"`
	CreatedAt     string    `json:"createdAt"`
	UpdatedAt     string    `json:"updatedAt"`
}
