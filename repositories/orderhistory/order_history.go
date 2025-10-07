package repositories

import (
	"context"
	"order-service/domain/dto"
	"order-service/domain/models"

	errWrap "order-service/common/error"
	errConstant "order-service/constants/error"

	"gorm.io/gorm"
)

type OrderHistoryRepository struct {
	db *gorm.DB
}

type IOrderHistoryRepository interface {
	Create(context.Context, *gorm.DB, *dto.OrderHistoryRequest) error
}

func NewOrderHistoryRepository(db *gorm.DB) IOrderHistoryRepository {
	return &OrderHistoryRepository{db: db}
}

func (o *OrderHistoryRepository) Create(ctx context.Context, tx *gorm.DB, param *dto.OrderHistoryRequest) error {
	orderHistory := models.OrderHistory{
		OrderID: param.OrderID,
		Status:  param.Status,
	}

	err := tx.WithContext(ctx).Create(&orderHistory).Error

	if err != nil {
		return errWrap.WrapError(errConstant.ErrSQLError)
	}

	return nil
}
