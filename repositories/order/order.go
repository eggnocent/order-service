package repositories

import (
	"context"
	"errors"
	"fmt"
	"order-service/domain/dto"
	"order-service/domain/models"
	"strconv"
	"time"

	errWrap "order-service/common/error"
	errConstant "order-service/constants/error"
	errOrder "order-service/constants/error/order"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

type IOrderRepository interface {
	FindAllWithPagination(context.Context, *dto.OrderRequestParam) ([]models.Order, int64, error)
	FindByUUID(context.Context, string) (*models.Order, error)
	FindByUserID(context.Context, string) ([]models.Order, error)
	Create(context.Context, *gorm.DB, *models.Order) (*models.Order, error)
	Update(context.Context, *gorm.DB, *models.Order, uuid.UUID) error
}

func NewOrderRepository(db *gorm.DB) IOrderRepository {
	return &OrderRepository{db: db}
}

func (o *OrderRepository) FindAllWithPagination(ctx context.Context, params *dto.OrderRequestParam) ([]models.Order, int64, error) {
	var orders []models.Order
	var sort string
	var total int64

	if params.SortColumn != nil {
		sort = fmt.Sprintf("%s %s", *params.SortColumn, *params.SortOrder)
	} else {
		sort = "created_at desc"
	}

	limit := params.Limit
	offset := (params.Page - 1) * params.Limit

	err := o.db.WithContext(ctx).Limit(limit).Offset(offset).Order(sort).Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	err = o.db.WithContext(ctx).Model(&models.Order{}).Count(&total).Error
	if err != nil {
		return nil, 0, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return orders, total, nil
}

func (o *OrderRepository) FindByUUID(ctx context.Context, orderUUID string) (*models.Order, error) {
	var order models.Order

	err := o.db.WithContext(ctx).Where("uuid = ?", orderUUID).First(&order).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errOrder.ErrOrderNotFound)
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return &order, nil
}

func (o *OrderRepository) FindByUserID(ctx context.Context, userID string) ([]models.Order, error) {
	var orders []models.Order

	err := o.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&orders).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errOrder.ErrOrderNotFound)
		}
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return orders, nil
}

func (o *OrderRepository) incrementCode(ctx context.Context) (*string, error) {
	var (
		order  *models.Order
		result string
		today  = time.Now().Format("20060102")
	)

	err := o.db.WithContext(ctx).Order("id desc").First(&order).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errWrap.WrapError(errConstant.ErrSQLError)
		}
	}

	if order.ID != 0 {
		orderCode := order.Code
		splitOrderName, _ := strconv.Atoi(orderCode[4:9])
		code := splitOrderName + 1
		result = fmt.Sprintf("ORD-%05d-%s", code, today)
	} else {
		result = fmt.Sprintf("ORD-%05d-%s", 1, today)
	}

	return &result, nil
}

func (o *OrderRepository) Create(ctx context.Context, tx *gorm.DB, param *models.Order) (*models.Order, error) {
	code, err := o.incrementCode(ctx)
	if err != nil {
		return nil, err
	}

	order := &models.Order{
		UUID:   uuid.New(),
		Code:   *code,
		UserID: param.UserID,
		Amount: param.Amount,
		Date:   param.Date,
		Status: param.Status,
		IsPaid: param.IsPaid,
	}

	err = tx.WithContext(ctx).Create(order).Error
	if err != nil {
		return nil, errWrap.WrapError(errConstant.ErrSQLError)
	}

	return order, nil
}

func (o *OrderRepository) Update(ctx context.Context, tx *gorm.DB, param *models.Order, orderUUID uuid.UUID) error {
	err := tx.WithContext(ctx).Model(&models.Order{}).Where("uuid = ?", orderUUID).Updates(param).Error
	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}

	return nil
}
