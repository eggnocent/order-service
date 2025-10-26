package services

import (
	"context"
	"fmt"
	"log"
	"order-service/clients"
	clientField "order-service/clients/field"
	clientPayment "order-service/clients/payment"
	clientUser "order-service/clients/user"
	"order-service/common/util"
	"order-service/constants"
	errOrder "order-service/constants/error/order"
	"order-service/domain/dto"
	"order-service/domain/models"
	"order-service/repositories"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderService struct {
	repository repositories.IRepositoryRegistry
	client     clients.IClientRegistry
}

type IOrderService interface {
	GetAllWithPagination(context.Context, *dto.OrderRequestParam) (*util.PaginationResult, error)
	GetByUUID(context.Context, string) (*dto.OrderResponse, error)
	GetOrdersByUserID(context.Context) ([]dto.OrderByUserIDResponse, error)
	Create(context.Context, *dto.OrderRequest) (*dto.OrderResponse, error)
	HandlePayment(context.Context, *dto.PaymentData) error
}

func NewOrderService(repo repositories.IRepositoryRegistry, client clients.IClientRegistry) IOrderService {
	return &OrderService{
		repository: repo,
		client:     client,
	}
}

func (o *OrderService) GetAllWithPagination(ctx context.Context, param *dto.OrderRequestParam) (*util.PaginationResult, error) {
	orders, total, err := o.repository.GetOrder().FindAllWithPagination(ctx, param)
	if err != nil {
		return nil, err
	}

	orderResult := make([]dto.OrderResponse, len(orders))
	for _, order := range orders {
		user, err := o.client.GetUser().GetUserbyUUID(ctx, order.UserID)
		if err != nil {
			return nil, err
		}

		orderResult = append(orderResult, dto.OrderResponse{
			UUID:      order.UUID,
			Code:      order.Code,
			UserName:  user.Name,
			Amount:    order.Amount,
			Status:    order.Status.GetStatusString(),
			OrderDate: order.Date,
			CreatedAt: *order.CreatedAt,
			UpdatedAt: *order.UpdatedAt,
		})
	}

	pagination := util.PaginationParams{
		Page:  param.Page,
		Limit: param.Limit,
		Count: total,
		Data:  orderResult,
	}

	response := util.GeneratePagination(pagination)
	return &response, nil
}

func (o *OrderService) GetByUUID(ctx context.Context, orderUUID string) (*dto.OrderResponse, error) {
	var (
		order *models.Order
		user  *clientUser.UserData
		err   error
	)

	order, err = o.repository.GetOrder().FindByUUID(ctx, orderUUID)
	if err != nil {
		return nil, err
	}

	user, err = o.client.GetUser().GetUserbyUUID(ctx, order.UserID)
	if err != nil {
		return nil, err
	}

	response := dto.OrderResponse{
		UUID:      order.UUID,
		Code:      order.Code,
		UserName:  user.Name,
		Amount:    order.Amount,
		Status:    order.Status.GetStatusString(),
		OrderDate: order.Date,
		CreatedAt: *order.CreatedAt,
		UpdatedAt: *order.UpdatedAt,
	}

	return &response, nil
}

func (o *OrderService) GetOrdersByUserID(ctx context.Context) ([]dto.OrderByUserIDResponse, error) {
	var (
		order []models.Order
		user  = ctx.Value(constants.User).(*clientUser.UserData)
		err   error
	)

	order, err = o.repository.GetOrder().FindByUserID(ctx, user.UUID.String())
	if err != nil {
		return nil, err
	}

	orderResult := make([]dto.OrderByUserIDResponse, 0, len(order))
	for _, ord := range order {
		payment, err := o.client.GetPayment().GetPaymentByUUID(ctx, ord.PaymentID)
		if err != nil {
			return nil, err
		}

		orderResult = append(orderResult, dto.OrderByUserIDResponse{
			Code:        ord.Code,
			Amount:      fmt.Sprintf("%s", util.RupiahFormat(&ord.Amount)),
			Status:      ord.Status.GetStatusString(),
			OrderDate:   ord.Date,
			PaymentLink: payment.PaymentLink,
			InvoiceLink: payment.InvoiceLink,
		})
	}

	return orderResult, nil
}

func (o *OrderService) Create(ctx context.Context, param *dto.OrderRequest) (*dto.OrderResponse, error) {
	var (
		order               *models.Order
		txErr, err          error
		user                = ctx.Value(constants.User).(*clientUser.UserData)
		field               *clientField.FieldData
		paymentResponse     *clientPayment.PaymentData
		orderFieldSchedules = make([]models.OrderField, 0, len(param.FieldScheduleIDs))
		totalAmount         float64
	)

	log.Printf("🟢 Create order request: %+v\n", param)
	log.Printf("👤 User from context: %+v\n", user)

	for _, fieldID := range param.FieldScheduleIDs {
		log.Printf("🔎 Fetching field data for UUID: %s\n", fieldID)
		uuidParsed := uuid.MustParse(fieldID)

		field, err = o.client.GetField().GetFieldByUUID(ctx, uuidParsed)
		if err != nil {
			log.Printf("❌ Error fetching field %s: %v\n", fieldID, err)
			return nil, err
		}

		if field.PricePerHour <= 0 {
			log.Printf("❌ Field %s has invalid price: %.2f", field.UUID, field.PricePerHour)
			return nil, fmt.Errorf("invalid price for field: %s", field.UUID)
		}

		if strings.TrimSpace(field.FieldName) == "" {
			log.Printf("❌ Field name is empty for field %s\n", field.UUID)
			return nil, fmt.Errorf("field name cannot be empty for field %s", field.UUID)
		}

		log.Printf("✅ Field data: %+v\n", field)

		if field.Status == constants.BookedStatus.String() {
			log.Printf("🚫 Field %s already booked\n", fieldID)
			return nil, errOrder.ErrFieldAlreadyBooked
		}

		totalAmount += field.PricePerHour
	}

	log.Printf("💰 Total order amount: %.2f\n", totalAmount)

	if strings.TrimSpace(user.PhoneNumber) == "" {
		log.Printf("❌ Phone number is empty for user: %s\n", user.UUID)
		return nil, fmt.Errorf("user phone number is required")
	}

	err = o.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		log.Println("🚧 Starting DB transaction")

		order, txErr = o.repository.GetOrder().Create(ctx, tx, &models.Order{
			UserID: user.UUID,
			Amount: totalAmount,
			Date:   time.Now(),
			Status: constants.Pending,
			IsPaid: false,
		})
		if txErr != nil {
			log.Printf("❌ Failed to create order: %v\n", txErr)
			return txErr
		}
		log.Printf("✅ Created order: %+v\n", order)

		for _, fieldID := range param.FieldScheduleIDs {
			uuidParsed := uuid.MustParse(fieldID)
			orderFieldSchedules = append(orderFieldSchedules, models.OrderField{
				OrderID:         order.ID,
				FieldScheduleID: uuidParsed,
			})
		}

		log.Printf("📌 Creating order-field schedule relation: %+v\n", orderFieldSchedules)
		txErr = o.repository.GetOrderField().Create(ctx, tx, orderFieldSchedules)
		if txErr != nil {
			log.Printf("❌ Failed to create order-field schedule: %v\n", txErr)
			return txErr
		}

		log.Println("📌 Creating order history")
		txErr = o.repository.GetOrderHistory().Create(ctx, tx, &dto.OrderHistoryRequest{
			Status:  constants.Pending.GetStatusString(),
			OrderID: order.ID,
		})
		if txErr != nil {
			log.Printf("❌ Failed to create order history: %v\n", txErr)
			return txErr
		}

		expiredAt := time.Now().Add(1 * time.Hour)
		description := fmt.Sprintf("Pembayaran sewa %s", field.FieldName)

		// 🔍 Buat dan log payload payment
		paymentRequest := &dto.PaymentRequest{
			OrderID:     order.UUID,
			ExpiredAt:   time.Unix(expiredAt.Unix(), 0),
			Amount:      order.Amount,
			Description: description,
			CustomerDetail: dto.CustomerDetail{
				Name:  user.Name,
				Email: user.Email,
				Phone: user.PhoneNumber,
			},
			ItemDetails: []dto.ItemDetails{
				{
					ID:       uuid.New(),
					Name:     description,
					Amount:   totalAmount,
					Quantity: 1,
				},
			},
		}

		log.Printf("📤 Payment Request Payload:\nOrderID: %s\nExpiredAt: %s\nAmount: %.2f\nDescription: %q\nCustomer: %s / %s / %s\nItems: %+v\n",
			paymentRequest.OrderID,
			paymentRequest.ExpiredAt.Format(time.RFC3339),
			paymentRequest.Amount,
			paymentRequest.Description,
			paymentRequest.CustomerDetail.Name,
			paymentRequest.CustomerDetail.Email,
			paymentRequest.CustomerDetail.Phone,
			paymentRequest.ItemDetails,
		)

		// 🔗 Kirim request payment link
		paymentResponse, txErr = o.client.GetPayment().CreatePaymentLink(ctx, paymentRequest)
		if txErr != nil {
			log.Printf("❌ Failed to create payment link: %v\n", txErr)
			return txErr
		}

		log.Printf("✅ Payment link created: %+v\n", paymentResponse)

		log.Println("🔄 Updating order with payment UUID")
		txErr = o.repository.GetOrder().Update(ctx, tx, &models.Order{
			PaymentID: paymentResponse.UUID,
		}, order.UUID)
		if txErr != nil {
			log.Printf("❌ Failed to update order with payment ID: %v\n", txErr)
			return txErr
		}

		log.Println("✅ Transaction committed successfully")
		return nil
	})

	if err != nil {
		log.Printf("❌ Error in Create order transaction: %v\n", err)
		return nil, err
	}

	response := &dto.OrderResponse{
		UUID:        order.UUID,
		Code:        order.Code,
		UserName:    user.Name,
		Amount:      order.Amount,
		Status:      order.Status.GetStatusString(),
		PaymentLink: paymentResponse.PaymentLink,
		OrderDate:   order.Date,
		CreatedAt:   *order.CreatedAt,
		UpdatedAt:   *order.UpdatedAt,
	}

	log.Printf("✅ Final Order Response: %+v\n", response)
	return response, nil
}

func (o *OrderService) mapPaymentStatusToOrder(request *dto.PaymentData) (constants.OrderStatus, *models.Order) {
	var (
		status constants.OrderStatus
		order  *models.Order
	)

	// apa yang sedang diikirm dari payment service di kafka
	switch request.Status {
	case constants.SettlementPaymentStatus:
		status = constants.PaymentSuccess
		order = &models.Order{
			IsPaid:    true,
			PaymentID: request.PaymentID,
			PaidAt:    request.PaidAt,
			Status:    status,
		}
	case constants.ExpiredPaymentStatus:
		status = constants.Expired
		order = &models.Order{
			IsPaid:    false,
			PaymentID: request.PaymentID,
			Status:    status,
		}
	case constants.PendingPaymentStatus:
		status = constants.PendingPayment
		order = &models.Order{
			IsPaid:    false,
			PaymentID: request.PaymentID,
			Status:    status,
		}
	}
	return status, order
}

func (o *OrderService) HandlePayment(ctx context.Context, request *dto.PaymentData) error {
	var (
		err, txErr          error
		order               *models.Order
		orderFieldSchedules []models.OrderField
	)
	status, body := o.mapPaymentStatusToOrder(request)
	err = o.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		txErr = o.repository.GetOrder().Update(ctx, tx, body, request.OrderID)
		if txErr != nil {
			return txErr
		}

		order, txErr = o.repository.GetOrder().FindByUUID(ctx, request.OrderID.String())
		if txErr != nil {
			return txErr
		}

		txErr = o.repository.GetOrderHistory().Create(ctx, tx, &dto.OrderHistoryRequest{
			Status:  status.GetStatusString(),
			OrderID: order.ID,
		})

		if request.Status == constants.SettlementPaymentStatus {
			orderFieldSchedules, txErr = o.repository.GetOrderField().FindByOrderID(ctx, order.ID)
			if txErr != nil {
				return txErr
			}

			filedScheduleIDs := make([]string, 0, len(orderFieldSchedules))
			for _, item := range orderFieldSchedules {
				filedScheduleIDs = append(filedScheduleIDs, item.FieldScheduleID.String())
			}

			txErr = o.client.GetField().UpdateStatus(&dto.UpdateFieldScheduleStatusRequest{
				FieldScheduleIDs: filedScheduleIDs,
			})
			if txErr != nil {
				return txErr
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
