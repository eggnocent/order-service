package services

import (
	"order-service/clients"
	"order-service/repositories"
	services "order-service/services/order"
)

type Registry struct {
	repository repositories.IRepositoryRegistry
	client     clients.IClientRegistry
}

type IServiceRegistry interface {
	GetOrder() services.IOrderService
}

func NewServiceRegistry(repo repositories.IRepositoryRegistry, client clients.IClientRegistry) IServiceRegistry {
	return &Registry{
		repository: repo,
		client:     client,
	}
}

func (r *Registry) GetOrder() services.IOrderService {
	return services.NewOrderService(r.repository, r.client)
}
