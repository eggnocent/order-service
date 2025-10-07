package controllers

import (
	controllers "order-service/controllers/http/order"
	"order-service/services"
)

type Registry struct {
	services services.IServiceRegistry
}

type IControllerRegistry interface {
	GetOrder() controllers.IOrderController
}

func NewControllerRegistry(services services.IServiceRegistry) *Registry {
	return &Registry{
		services: services,
	}
}

func (r *Registry) GetOrder() controllers.IOrderController {
	return controllers.NewOrderController(r.services)
}
