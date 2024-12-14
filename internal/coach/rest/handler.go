package rest

import (
	coachGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.coach"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
)

type Handler struct {
	coachClient *coachGRPC.CoachClient
}

func NewHandler(coachClient *coachGRPC.CoachClient) *Handler {
	return &Handler{
		coachClient: coachClient,
	}
}

func (h *Handler) GetCoaches(c *gin.Context) {
	coaches, err := (*h.coachClient).GetCoachesWithServices(c.Request.Context(), &emptypb.Empty{})
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coaches": coaches,
	})
}
