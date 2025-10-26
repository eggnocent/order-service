package controllers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"order-service/common/response"
	"order-service/domain/dto"
	"order-service/services"

	error2 "order-service/common/error"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrderController struct {
	service services.IServiceRegistry
}

type IOrderController interface {
	GetAllWthPagination(ctx *gin.Context)
	GetByUUID(ctx *gin.Context)
	GetOrderByUserID(ctx *gin.Context)
	Create(ctx *gin.Context)
}

func NewOrderController(service services.IServiceRegistry) *OrderController {
	return &OrderController{
		service: service,
	}
}

func (c *OrderController) GetAllWthPagination(ctx *gin.Context) {
	var params dto.OrderRequestParam
	err := ctx.ShouldBindQuery(&params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	validate := validator.New()
	if err = validate.Struct(params); err != nil {
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := error2.ErrValidationResponse(err)
		response.HttpResponse(response.ParamHTTPResp{
			Err:     err,
			Code:    http.StatusUnprocessableEntity,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     ctx,
		})
		return
	}

	result, err := c.service.GetOrder().GetAllWithPagination(ctx, &params)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusInternalServerError,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  ctx,
	})
}

func (c *OrderController) GetByUUID(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	result, err := c.service.GetOrder().GetByUUID(ctx, uuid)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusInternalServerError,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  ctx,
	})
}

func (c *OrderController) GetOrderByUserID(ctx *gin.Context) {
	result, err := c.service.GetOrder().GetOrdersByUserID(ctx)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusInternalServerError,
			Err:  err,
			Gin:  ctx,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusOK,
		Data: result,
		Gin:  ctx,
	})
}

func (o *OrderController) Create(c *gin.Context) {
	var (
		request dto.OrderRequest
		ctx     = c.Request.Context()
	)

	// Log request headers
	for k, v := range c.Request.Header {
		log.Printf("📥 Header [%s] = %v\n", k, v)
	}

	// Log request body secara mentah (gunakan io.ReadAll untuk debugging)
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	log.Printf("📥 Raw Body: %s\n", string(bodyBytes))

	// Reset Body agar bisa dibaca lagi oleh ShouldBindJSON
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Bind JSON
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Printf("❌ Gagal bind JSON: %v\n", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	log.Printf("✅ Berhasil bind JSON: %+v\n", request)

	// Validasi
	validate := validator.New()
	if err = validate.Struct(request); err != nil {
		log.Printf("❌ Validasi gagal: %v\n", err)
		errMessage := http.StatusText(http.StatusUnprocessableEntity)
		errorResponse := error2.ErrValidationResponse(err)

		response.HttpResponse(response.ParamHTTPResp{
			Err:     err,
			Code:    http.StatusUnprocessableEntity,
			Message: &errMessage,
			Data:    errorResponse,
			Gin:     c,
		})
		return
	}

	// Panggil service
	result, err := o.service.GetOrder().Create(ctx, &request)
	if err != nil {
		log.Printf("❌ Gagal create order: %v\n", err)
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	log.Printf("✅ Order berhasil dibuat: %+v\n", result)

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Data: result,
		Gin:  c,
	})
}
