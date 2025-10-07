package repositories

import (
	orderRepo "order-service/repositories/order"
	orderFieldRepo "order-service/repositories/orderfield"
	orderHistoryRepo "order-service/repositories/orderhistory"

	"gorm.io/gorm"
)

type Registry struct {
	db *gorm.DB
}

type IRepositoryRegistry interface {
	GetOrder() orderRepo.IOrderRepository
	GetOrderField() orderFieldRepo.IOrderFieldRepository
	GetOrderHistory() orderHistoryRepo.IOrderHistoryRepository
	GetTx() *gorm.DB
}

func NewRepositoryRegistry(db *gorm.DB) IRepositoryRegistry {
	return &Registry{db: db}
}

func (r *Registry) GetOrder() orderRepo.IOrderRepository {
	return orderRepo.NewOrderRepository(r.db)
}

func (r *Registry) GetOrderField() orderFieldRepo.IOrderFieldRepository {
	return orderFieldRepo.NewOrderFieldRepository(r.db)
}

func (r *Registry) GetOrderHistory() orderHistoryRepo.IOrderHistoryRepository {
	return orderHistoryRepo.NewOrderHistoryRepository(r.db)
}

func (r *Registry) GetTx() *gorm.DB {
	return r.db.Begin()
}
