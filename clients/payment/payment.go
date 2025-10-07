package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"order-service/clients/config"
	"order-service/common/util"
	configApp "order-service/config"
	"order-service/constants"
	"order-service/domain/dto"
	"time"

	"github.com/google/uuid"
)

type PaymentClient struct {
	client config.IClientConfig
}

type IPaymentClient interface {
	GetPaymentByUUID(context.Context, uuid.UUID) (*PaymentData, error)
	CreatePaymentLink(context.Context, *dto.PaymentRequest) (*PaymentData, error)
}

func NewPaymentClient(client config.IClientConfig) IPaymentClient {
	return &PaymentClient{client: client}
}

func (p *PaymentClient) GetPaymentByUUID(ctx context.Context, paymentUUID uuid.UUID) (*PaymentData, error) {
	unixTime := time.Now().Unix()
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		configApp.Config.AppName,
		p.client.SignatureKey(),
		unixTime,
	)

	apiKey := util.GenerateSHA256(generateAPIKey)
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	var response PaymentResponse
	request := p.client.Client().Clone().
		Set(constants.Authorization, bearerToken).
		Set(constants.XServiceName, configApp.Config.AppName).
		Set(constants.XServiceName, apiKey).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Get(fmt.Sprintf("%s/api/v1/payments/%s", p.client.BaseURL(), paymentUUID))

	resp, _, errrs := request.EndStruct(&response)

	if len(errrs) > 0 {
		return nil, errrs[0]
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment response: %s", response.Message)
	}

	data, ok := response.Data.(PaymentData)
	if !ok {
		return nil, fmt.Errorf("failed to cast response data to PaymentData")
	}

	return &data, nil
}

func (p *PaymentClient) CreatePaymentLink(ctx context.Context, req *dto.PaymentRequest) (*PaymentData, error) {
	unixTime := time.Now().Unix()
	generateAPIKey := fmt.Sprintf("%s:%s:%d",
		configApp.Config.AppName,
		p.client.SignatureKey(),
		unixTime,
	)

	apiKey := util.GenerateSHA256(generateAPIKey)
	token := ctx.Value(constants.Token).(string)
	bearerToken := fmt.Sprintf("Bearer %s", token)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, bodyResp, errs := p.client.Client().Clone().
		Post(fmt.Sprintf("%s/api/v1/payments", p.client.BaseURL())).
		Set(constants.Authorization, bearerToken).
		Set(constants.XServiceName, configApp.Config.AppName).
		Set(constants.XApiKey, apiKey).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Send(string(body)).
		End()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	var response PaymentResponse
	if resp.StatusCode != http.StatusCreated {
		err = json.Unmarshal([]byte(bodyResp), &response)
		if err != nil {
			return nil, err
		}

		paymentError := fmt.Errorf("payment response: %s", response.Message)
		return nil, paymentError
	}

	err = json.Unmarshal([]byte(bodyResp), &response)
	if err != nil {
		return nil, err
	}

	// Fix: Gunakan type assertion seperti di GetPaymentByUUID
	data, ok := response.Data.(PaymentData)
	if !ok {
		return nil, fmt.Errorf("failed to cast response data to PaymentData")
	}

	return &data, nil
}
