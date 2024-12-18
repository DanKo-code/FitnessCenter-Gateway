package rest

import (
	abonementGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.abonement"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.Engine, abonementClient *abonementGRPC.AbonementClient) {
	h := NewHandler(abonementClient)

	router.GET("/abonements", h.GetAbonements)
	router.POST("/abonements", h.CreateAbonement)
	router.PUT("/abonements", h.UpdateAbonement)
	router.DELETE("/abonements/:id", h.DeleteAbonement)
}
