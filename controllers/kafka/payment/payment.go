package kafka

import (
	"context"
	"encoding/json"
	"order-service/common/util"
	"order-service/domain/dto"
	"order-service/services"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

const PaymentTopic = "paymet-service-callback"

type PaymentKafka struct {
	service services.IServiceRegistry
}

type IPaymentKafka interface {
	HandlePayment(context.Context, *sarama.ConsumerMessage) error
}

func NewPaymentKafka(service services.IServiceRegistry) IPaymentKafka {
	return &PaymentKafka{
		service: service,
	}
}

func (p *PaymentKafka) HandlePayment(ctx context.Context, msg *sarama.ConsumerMessage) error {
	defer util.Recover()
	var body dto.PaymentContent

	err := json.Unmarshal(msg.Value, &body)
	if err != nil {
		logrus.Error("[PaymentKafka-HandlePayment] error when unmarshal message: ", err)
		return err
	}

	data := body.Body.Data
	err = p.service.GetOrder().HandlePayment(ctx, &data)
	if err != nil {
		logrus.Error("[PaymentKafka-HandlePayment] error when handle payment: ", err)
		return err
	}

	logrus.Info("[PaymentKafka-HandlePayment] success handle payment: ", data)
	return nil

}
