package config

import (
	"order-service/common/util"
	"os"

	"github.com/sirupsen/logrus"
)

var Config AppConfig

type AppConfig struct {
	Port                  int             `json:"port"`
	AppName               string          `json:"appName"`
	AppEnv                string          `json:"appEnv"`
	SignatureKey          string          `json:"signatureKey"`
	Database              Database        `json:"database"`
	RateLimiterMaxRequest float64         `json:"rateLimiterMaxRequest"`
	RateLimiterTimeSecond int             `json:"rateLimiterTimeSecond"`
	InternalService       InternalService `json:"internalService"`
	Kafka                 Kafka           `json:"kafka"`
}

type Database struct {
	Host                  string `json:"host"`
	Port                  int    `json:"port"`
	Name                  string `json:"name"`
	Username              string `json:"username"`
	Password              string `json:"password"`
	MaxOpenConnections    int    `json:"maxOpenConnections"`
	MaxLifeTimeConnection int    `json:"maxLifeTimeConnection"`
	MaxIdleConnections    int    `json:"maxIdleConnections"`
	MaxIdleTime           int    `json:"maxIdleTime"`
}

type InternalService struct {
	User    User    `json:"user"`
	Field   Field   `json:"field"`
	Payment Payment `json:"payment"`
}

type User struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type Field struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type Payment struct {
	Host         string `json:"host"`
	SignatureKey string `json:"signatureKey"`
}

type Kafka struct {
	Brokers               []string `json:"brokers"`
	TimeoutInMs           int      `json:"timeoutInMs"`
	MaxRetry              int      `json:"maxRetry"`
	MaxWaitTimeInMs       int      `json:"maxWaitTimeInMs"`
	MaxProcessingTimeInMs int      `json:"maxProcessingTimeInMs"`
	BackoffTimeInMs       int      `json:"backoffTimeInMs"`
	Topics                []string `json:"topics"`
	GroupID               string   `json:"groupID"`
}

func Init() {
	err := util.BindFromJSON(&Config, "config.json", ".")
	if err == nil {
		logrus.Info("loaded config from local file: config.json")
		return
	}

	logrus.Warnf("failed to bind config.json: %v", err)

	consulURL := os.Getenv("CONSUL_HTTP_URL")
	consulPath := os.Getenv("CONSUL_HTTP_PATH")
	if consulURL == "" || consulPath == "" {
		logrus.Fatal("config.json not found, and CONSUL_HTTP_URL or CONSUL_HTTP_PATH is not set")
	}

	logrus.Infof("attempting to load config from Consul: %s/%s", consulURL, consulPath)
	err = util.BindFromConsul(&Config, consulURL, consulPath)
	if err != nil {
		logrus.Fatalf("failed to bind config from Consul: %v", err)
	}

	logrus.Info("loaded config from Consul")
}
