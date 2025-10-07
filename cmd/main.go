package cmd

import (
	"context"
	"fmt"
	"net/http"
	"order-service/clients"
	"order-service/common/response"
	"order-service/config"
	"order-service/constants"
	controllers "order-service/controllers/http"
	kafka2 "order-service/controllers/kafka"
	kafka "order-service/controllers/kafka/config"
	"order-service/domain/models"
	"order-service/middlewares"
	"order-service/repositories"
	"order-service/routes"
	"order-service/services"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var command = &cobra.Command{
	Use:   "Serve",
	Short: "Run the HTTP server and Kafka consumer",
	Run: func(cmd *cobra.Command, args []string) {
		config.Init()

		db, err := config.InitDatabase()
		if err != nil {
			panic(err)
		}

		loc, err := time.LoadLocation("Asia/Jakarta")
		if err != nil {
			panic(err)
		}

		time.Local = loc

		err = db.AutoMigrate(
			&models.Order{},
			&models.OrderHistory{},
			&models.OrderField{},
		)

		client := clients.NewClientRegistry()
		repository := repositories.NewRepositoryRegistry(db)
		service := services.NewServiceRegistry(repository, client)
		controller := controllers.NewControllerRegistry(service)

		serveHttp(controller, client)
		serveKafkaConsumer(service)
	},
}

func Run() {
	if err := command.Execute(); err != nil {
		panic(err)
	}
}

func serveHttp(controllers controllers.IControllerRegistry, client clients.IClientRegistry) {
	router := gin.Default()
	router.Use(middlewares.HandlePanic())
	router.Use(gin.Logger())

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, response.Response{
			Status:  constants.Error,
			Message: fmt.Sprintf("Path %s", http.StatusText(http.StatusNotFound)),
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, response.Response{
			Status:  constants.Success,
			Message: "Welcome to Payment Service",
		})
	})

	// CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, PATCH")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-service-name, x-api-key, x-request-at")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Optional rate limiter (komentar masih kamu simpan)
	// lmt := tollbooth.NewLimiter(
	// 	config.Config.RateLimiterMaxRequests,
	// 	&limiter.ExpirableOptions{
	// 		DefaultExpirationTTL: payment.Duration(config.Config.RateLimiterTimeSeconds) * payment.Second,
	// 	})
	// router.Use(middlewares.RateLimiter(lmt))

	group := router.Group("/api/v1")
	route := routes.NewRouteRegistry(group, controllers, client)
	route.Serve()

	go func() {
		port := fmt.Sprintf(":%d", config.Config.Port)
		router.Run(port)
	}()
}

func serveKafkaConsumer(service services.IServiceRegistry) {
	kafkaConsumerConfig := sarama.NewConfig()
	kafkaConsumerConfig.Consumer.MaxWaitTime = time.Duration(config.Config.Kafka.MaxWaitTimeInMs) * time.Millisecond
	kafkaConsumerConfig.Consumer.MaxProcessingTime = time.Duration(config.Config.Kafka.MaxProcessingTimeInMs) * time.Millisecond
	kafkaConsumerConfig.Consumer.Retry.Backoff = time.Duration(config.Config.Kafka.BackoffTimeInMs) * time.Millisecond
	kafkaConsumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	kafkaConsumerConfig.Consumer.Offsets.AutoCommit.Enable = true
	kafkaConsumerConfig.Consumer.Offsets.AutoCommit.Interval = 1 * time.Second
	kafkaConsumerConfig.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	brokers := config.Config.Kafka.Brokers
	groupID := config.Config.Kafka.GroupID
	topic := config.Config.Kafka.Topics

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, kafkaConsumerConfig)
	if err != nil {
		logrus.Fatalf("Error creating Kafka consumer group: %v", err)
		return
	}

	defer consumerGroup.Close()

	consumer := kafka.NewConsumerGroup()
	kafkaRegistry := kafka2.NewKafkaRegistry(service)
	kafkaConsumer := kafka.NewKafkaConsumer(consumer, kafkaRegistry)
	kafkaConsumer.Register()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	ctx, cancle := context.WithCancel(context.Background())
	defer cancle()

	go func() {
		for {
			err = consumerGroup.Consume(ctx, topic, consumer)
			if err != nil {
				logrus.Errorf("Error consuming Kafka messages: %v", err)
				panic(err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	logrus.Infof("Kafka consumer up and running! Listening to topics: %v", topic)

	<-signals
	logrus.Info("Termination signal received. Shutting down Kafka consumer...")
}
