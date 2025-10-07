package clients

import (
	"order-service/clients/config"
	clients3 "order-service/clients/field"
	clients2 "order-service/clients/payment"
	clients "order-service/clients/user"
	config2 "order-service/config"
)

type ClientRegistry struct {
}

type IClientRegistry interface {
	GetUser() clients.IUserClient
	GetPayment() clients2.IPaymentClient
	GetField() clients3.IFieldClient
}

func NewClientRegistry() IClientRegistry {
	return &ClientRegistry{}
}

func (c *ClientRegistry) GetUser() clients.IUserClient {
	return clients.NewUserClient(
		config.NewClientConfig(
			config.WithBaseURL(config2.Config.InternalService.User.Host),
			config.WithSignatureKey(config2.Config.InternalService.User.SignatureKey),
		))
}

func (c *ClientRegistry) GetPayment() clients2.IPaymentClient {
	return clients2.NewPaymentClient(
		config.NewClientConfig(
			config.WithBaseURL(config2.Config.InternalService.Payment.Host),
			config.WithSignatureKey(config2.Config.InternalService.Payment.SignatureKey),
		))
}

func (c *ClientRegistry) GetField() clients3.IFieldClient {
	return clients3.NewFieldClient(
		config.NewClientConfig(
			config.WithBaseURL(config2.Config.InternalService.Field.Host),
			config.WithSignatureKey(config2.Config.InternalService.Field.SignatureKey),
		))
}
