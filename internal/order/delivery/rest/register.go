package rest

import (
	orderGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.order"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.Engine, orderClient *orderGRPC.OrderClient) {
	h := NewHandler(orderClient)

	router.POST("/checkout-session-completed", h.HandleCheckoutSessionCompleted)
	router.POST("/create-checkout-session", h.CreateCheckoutSession)
	router.GET("/orders/:userId", h.GetUserOrders)
}
