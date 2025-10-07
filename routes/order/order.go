package routes

import (
	"order-service/clients"
	"order-service/constants"
	controllers "order-service/controllers/http"
	"order-service/middlewares"

	"github.com/gin-gonic/gin"
)

type OrderRoutes struct {
	controller controllers.IControllerRegistry
	client     clients.IClientRegistry
	group      *gin.RouterGroup
}

type IOrderRoute interface {
	Run()
}

func NewOrderRoutes(group *gin.RouterGroup, controller controllers.IControllerRegistry, client clients.IClientRegistry) IOrderRoute {
	return &OrderRoutes{
		controller: controller,
		client:     client,
		group:      group,
	}
}

func (r *OrderRoutes) Run() {
	group := r.group.Group("/order")
	group.Use(middlewares.Authenticate())

	group.GET("", middlewares.CheckRole([]string{constants.Admin, constants.Customer}, r.client), r.controller.GetOrder().GetAllWthPagination)
	group.GET("/:uuid", middlewares.CheckRole([]string{constants.Admin, constants.Customer}, r.client), r.controller.GetOrder().GetByUUID)
	group.GET("/user", middlewares.CheckRole([]string{constants.Customer}, r.client), r.controller.GetOrder().GetOrderByUserID)
	group.POST("", middlewares.CheckRole([]string{constants.Customer}, r.client), r.controller.GetOrder().Create)
}
