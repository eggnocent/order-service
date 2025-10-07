package services

import (
	"context"
	"fmt"
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
		paymentResponse     *clientPayment.PaymentData // Fix: Ganti PaymentResponse ke PaymentData
		orderFieldSchedules = make([]models.OrderField, 0, len(param.FieldScheduleIDs))
		totalAmount         float64
	)

	for _, fieldID := range param.FieldScheduleIDs {
		uuidParsed := uuid.MustParse(fieldID)
		field, err = o.client.GetField().GetFieldByUUID(ctx, uuidParsed)
		if err != nil {
			return nil, err
		}

		totalAmount += field.PricePerHour
		if field.Status == constants.BookedStatus.String() {
			return nil, errOrder.ErrFieldAlreadyBooked
		}
	}

	err = o.repository.GetTx().Transaction(func(tx *gorm.DB) error {
		order, txErr = o.repository.GetOrder().Create(ctx, tx, &models.Order{
			UserID: user.UUID,
			Amount: totalAmount,
			Date:   time.Now(),
			Status: constants.Pending,
			IsPaid: false,
		})
		if txErr != nil {
			return txErr
		}

		for _, fieldID := range param.FieldScheduleIDs {
			uuidParsed := uuid.MustParse(fieldID)
			orderFieldSchedules = append(orderFieldSchedules, models.OrderField{
				OrderID:         order.ID,
				FieldScheduleID: uuidParsed,
			})
		}

		txErr = o.repository.GetOrderField().Create(ctx, tx, orderFieldSchedules)
		if txErr != nil {
			return txErr
		}

		txErr = o.repository.GetOrderHistory().Create(ctx, tx, &dto.OrderHistoryRequest{
			Status:  constants.Pending.GetStatusString(),
			OrderID: order.ID,
		})
		if txErr != nil {
			return txErr
		}

		expiredAt := time.Now().Add(1 * time.Hour) // set expired untuk paymentlink 1 jam dari ketika create order
		description := fmt.Sprintf("Pembayaran sewa %s", field.FieldName)
		paymentResponse, txErr = o.client.GetPayment().CreatePaymentLink(ctx, &dto.PaymentRequest{
			OrderID:     order.UUID,
			ExpiredAt:   expiredAt,
			Amount:      order.Amount,
			Description: description,
			CustomerDetail: dto.CustomerDetail{
				Name:  user.Name,
				Email: user.Email,
				Phone: user.PhoneNumber,
			},
			ItemsDetails: []dto.ItemDetails{
				{
					ID:       uuid.New(),
					Name:     description,
					Amount:   totalAmount,
					Quantity: 1,
				},
			},
		})

		if txErr != nil {
			return txErr
		}

		txErr = o.repository.GetOrder().Update(ctx, tx, &models.Order{
			PaymentID: paymentResponse.UUID,
		}, order.UUID)
		if txErr != nil {
			return txErr
		}
		return nil
	})

	if err != nil {
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
