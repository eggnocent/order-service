package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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

	// Log payload sebelum marshal
	log.Printf("📤 Outgoing Payment Request struct: %+v\n", req)

	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("❌ Error marshal payment request: %v\n", err)
		return nil, err
	}
	log.Printf("📦 JSON Payload: %s\n", string(body))

	// Execute request
	resp, bodyResp, errs := p.client.Client().Clone().
		Post(fmt.Sprintf("%s/api/v1/payments", p.client.BaseURL())).
		Set(constants.Authorization, bearerToken).
		Set(constants.XServiceName, configApp.Config.AppName).
		Set(constants.XApiKey, apiKey).
		Set(constants.XRequestAt, fmt.Sprintf("%d", unixTime)).
		Set("Content-Type", "application/json"). // Important!
		Send(string(body)).
		End()

	// Log error jika ada
	if len(errs) > 0 {
		log.Printf("❌ Resty Errors: %+v\n", errs)
		return nil, errs[0]
	}

	// Log status dan response raw
	log.Printf("📥 Status code from payment-service: %d\n", resp.StatusCode)
	log.Printf("📥 Raw response body: %s\n", bodyResp)

	var response PaymentResponse
	err = json.Unmarshal([]byte(bodyResp), &response)
	if err != nil {
		log.Printf("❌ Failed to unmarshal payment response: %v\n", err)
		return nil, err
	}
	log.Printf("📥 Parsed response: %+v\n", response)

	if resp.StatusCode != http.StatusCreated {
		paymentError := fmt.Errorf("payment response: %s", response.Message)
		return nil, paymentError
	}

	// Cek dan cast data
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		log.Printf("❌ Failed to cast response.Data to map[string]interface{}: %+v\n", response.Data)
		return nil, fmt.Errorf("failed to cast response data to PaymentData")
	}

	// Convert map[string]interface{} ke JSON → struct
	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("❌ Failed to marshal payment response data: %v\n", err)
		return nil, err
	}

	var paymentData PaymentData
	err = json.Unmarshal(dataBytes, &paymentData)
	if err != nil {
		log.Printf("❌ Failed to unmarshal payment data into struct: %v\n", err)
		return nil, err
	}

	log.Printf("✅ Final PaymentData parsed: %+v\n", paymentData)
	return &paymentData, nil
}
