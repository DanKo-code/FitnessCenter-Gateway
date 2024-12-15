package rest

import (
	reviewGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.review"
	userGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.user"
	"github.com/gin-gonic/gin"
)

func RegisterHTTPEndpoints(
	router *gin.Engine,
	reviewClient *reviewGRPC.ReviewClient,
	userClient *userGRPC.UserClient,
) {
	h := NewHandler(reviewClient, userClient)

	router.POST("/reviews", h.CreateCoachReview)
}
