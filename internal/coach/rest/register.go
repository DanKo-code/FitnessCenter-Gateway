package rest

import (
	coachGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.coach"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(router *gin.Engine, coachClient *coachGRPC.CoachClient) {
	h := NewHandler(coachClient)

	router.GET("/coaches", h.GetCoaches)
	router.POST("/coaches", h.CreateCoach)
	router.PUT("/coaches", h.UpdateCoach)
	router.DELETE("/coaches/:id", h.DeleteCoach)
}
