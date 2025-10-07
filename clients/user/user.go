package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"order-service/clients/config"
	"order-service/common/util"
	config2 "order-service/config"
	"order-service/constants"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type UserClient struct {
	client config.IClientConfig
}

type IUserClient interface {
	GetUserbyToken(ctx context.Context) (*UserData, error)
	GetUserbyUUID(context.Context, uuid.UUID) (*UserData, error)
}

func NewUserClient(client config.IClientConfig) IUserClient {
	return &UserClient{client: client}
}

func (u *UserClient) GetUserbyToken(ctx context.Context) (*UserData, error) {
	// 🔐 Generate secure headers
	unixTime := time.Now().Unix()
	requestAt := fmt.Sprintf("%d", unixTime)

	generateAPIKey := fmt.Sprintf("%s:%s:%s",
		config2.Config.AppName,
		u.client.SignatureKey(),
		requestAt,
	)

	apiKey := util.GenerateSHA256(generateAPIKey)

	logrus.Infof("🔏 [GetUserbyToken] generateAPIKey string: %s", generateAPIKey)
	logrus.Infof("🔏 [GetUserbyToken] apiKey (hashed): %s", apiKey)

	// 🔐 Ambil token dari context
	val := ctx.Value(constants.Token)
	logrus.Infof("🔍 [GetUserbyToken] Raw token value from context: %v", val)

	token, ok := val.(string)
	if !ok || token == "" {
		logrus.Warn("❌ [GetUserbyToken] TOKEN_NOT_FOUND_IN_CONTEXT or not string")
		return nil, errors.New("unauthorized")
	}

	bearerToken := fmt.Sprintf("Bearer %s", token)
	logrus.Infof("🔐 [GetUserbyToken] Bearer token: %s", bearerToken)

	// 🔧 Build request
	var response UserResponse
	request := u.client.Client().Clone().
		Set(constants.Authorization, bearerToken).
		Set(constants.XApiKey, apiKey).
		Set(constants.XServiceName, config2.Config.AppName).
		Set(constants.XRequestAt, requestAt)

	logrus.Infof("➡️ [GetUserbyToken] Sending request to Auth Service: %s/api/v1/auth/user", u.client.BaseURL())

	resp, _, errs := request.
		Get(fmt.Sprintf("%s/api/v1/auth/user", u.client.BaseURL())).
		EndStruct(&response)

	// 🔍 Handle response
	if len(errs) > 0 {
		logrus.Errorf("❌ [GetUserbyToken] HTTP error: %v", errs[0])
		return nil, errs[0]
	}

	logrus.Infof("⬅️ [GetUserbyToken] Response status code: %d", resp.StatusCode)
	logrus.Infof("⬅️ [GetUserbyToken] Response body message: %s", response.Message)

	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("🚫 [GetUserbyToken] Unauthorized - user response: %s", response.Message)
		return nil, fmt.Errorf("user response: %s", response.Message)
	}

	logrus.Infof("✅ [GetUserbyToken] User data retrieved successfully: %+v", response.Data)
	return &response.Data, nil
}

func (u *UserClient) GetUserbyUUID(ctx context.Context, uuid uuid.UUID) (*UserData, error) {
	unixTime := time.Now().Unix()
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		config2.Config.AppName,
		u.client.SignatureKey(),
		unixTime,
	)
	apiKey := util.GenerateSHA256(generateAPIKey)
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	var response UserResponse
	request := u.client.Client().Clone().
		Set(constants.Authorization, bearerToken).
		Set(constants.XServiceName, config2.Config.AppName).
		Set(constants.XApiKey, apiKey).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Get(fmt.Sprintf("%s/api/v1/auth/%s", u.client.BaseURL(), uuid))

	resp, _, errs := request.EndStruct(&response)
	if len(errs) > 0 {
		return nil, errs[0]
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user response: %s", response.Message)
	}

	return &response.Data, nil
}
