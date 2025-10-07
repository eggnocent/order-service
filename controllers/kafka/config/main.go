package kafka

import (
	"order-service/config"
	"order-service/controllers/kafka"

	kafka2 "order-service/controllers/kafka/payment"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type Kafka struct {
	consumer *ConsumerGroup
	kafka    kafka.IKafkaRegistry
}

type IKafkaInterface interface {
	Register()
}

func NewKafkaConsumer(consumer *ConsumerGroup, kafka kafka.IKafkaRegistry) IKafkaInterface {
	return &Kafka{consumer: consumer, kafka: kafka}
}

func (k *Kafka) Register() {
	k.paymentHandler()
}

func (k *Kafka) paymentHandler() {
	if slices.Contains(config.Config.Kafka.Topics, kafka2.PaymentTopic) {
		k.consumer.RegisterHandler(kafka2.PaymentTopic, k.kafka.GetPayment().HandlePayment)
		logrus.Infof("Payment handler registered for topic %s", kafka2.PaymentTopic)
	}
}
