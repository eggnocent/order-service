package kafka

import (
	"context"
	"order-service/config"
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type (
	TopicName string
	Handler   func(ctx context.Context, message *sarama.ConsumerMessage) error
)

type ConsumerGroup struct {
	handler map[TopicName]Handler
}

func NewConsumerGroup() *ConsumerGroup {
	return &ConsumerGroup{
		handler: make(map[TopicName]Handler),
	}
}

func (c *ConsumerGroup) Setup(sarama sarama.ConsumerGroupSession) error {
	logrus.Info("Kafka consumer group set up")
	return nil
}

func (c *ConsumerGroup) Cleanup(sarama sarama.ConsumerGroupSession) error {
	logrus.Info("Kafka consumer group clean up")
	return nil
}

func (c *ConsumerGroup) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	messages := claim.Messages()
	for message := range messages {
		handler, ok := c.handler[TopicName(message.Topic)]
		if !ok {
			logrus.Warnf("No handler for topic %s", message.Topic)
			continue
		}

		var err error
		maxRetry := config.Config.Kafka.MaxRetry
		for attempt := 1; attempt <= maxRetry; attempt++ {
			err = handler(context.Background(), message)
			if err == nil {
				break
			}

			logrus.Errorf("Error handling message from topic %s, attempt %d/%d: %v", message.Topic, attempt, maxRetry, err)
			if attempt == maxRetry {
				logrus.Errorf("Max retry reached for message from topic %s: %v", message.Topic, err)
			}
		}

		if err != nil {
			logrus.Errorf("Failed to process message from topic %s after %d attempts: %v", message.Topic, maxRetry, err)
			session.MarkMessage(message, err.Error())
			break
		}

		session.MarkMessage(message, time.Now().UTC().String())
	}

	return nil
}

func (c *ConsumerGroup) RegisterHandler(topic TopicName, handler Handler) {
	c.handler[topic] = handler
	logrus.Infof("Handler registered for topic %s", topic)
}
