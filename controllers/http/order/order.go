package controllers

import (
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

	err := c.ShouldBindJSON(&request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	validate := validator.New()
	if err = validate.Struct(request); err != nil {
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

	result, err := o.service.GetOrder().Create(ctx, &request)
	if err != nil {
		response.HttpResponse(response.ParamHTTPResp{
			Code: http.StatusBadRequest,
			Err:  err,
			Gin:  c,
		})
		return
	}

	response.HttpResponse(response.ParamHTTPResp{
		Code: http.StatusCreated,
		Data: result,
		Gin:  c,
	})
}
