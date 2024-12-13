package rest

import (
	abonementGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.abonement"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
)

type Handler struct {
	abonementClient *abonementGRPC.AbonementClient
}

func NewHandler(abonementClient *abonementGRPC.AbonementClient) *Handler {
	return &Handler{
		abonementClient: abonementClient,
	}
}

func (h *Handler) GetAbonements(c *gin.Context) {
	abonements, err := (*h.abonementClient).GetAbonementsWithServices(c.Request.Context(), &emptypb.Empty{})
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"abonements": abonements,
	})
}
