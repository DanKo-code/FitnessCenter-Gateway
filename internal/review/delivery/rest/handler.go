package rest

import (
	"Gateway/pkg/logger"
	"context"
	coachGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.coach"
	reviewGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.review"
	userGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.user"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct {
	reviewClient *reviewGRPC.ReviewClient
	userClient   *userGRPC.UserClient
}

func NewHandler(reviewClient *reviewGRPC.ReviewClient, userClient *userGRPC.UserClient) *Handler {
	return &Handler{
		reviewClient: reviewClient,
		userClient:   userClient,
	}
}

func (h *Handler) CreateCoachReview(c *gin.Context) {

	type CoachReviewDataForCreate struct {
		UserId  uuid.UUID
		Body    string
		CoachId uuid.UUID
	}

	coachReviewDataForCreate := &CoachReviewDataForCreate{}

	createCoachReviewRequest := &reviewGRPC.CreateCoachReviewRequest{
		ReviewDataForCreate: &reviewGRPC.CoachReviewDataForCreate{
			UserId:  "",
			Body:    "",
			CoachId: "",
		},
	}

	coachReviewDataForCreateProto := &reviewGRPC.CoachReviewDataForCreate{}

	if err := c.ShouldBindJSON(&coachReviewDataForCreate); err != nil {
		logger.ErrorLogger.Printf("Error binding CreateCoachReviewRequest: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coachReviewDataForCreateProto.CoachId = coachReviewDataForCreate.CoachId.String()
	coachReviewDataForCreateProto.Body = coachReviewDataForCreate.Body
	coachReviewDataForCreateProto.UserId = coachReviewDataForCreate.UserId.String()

	//CoachId validate
	_, err := uuid.Parse(coachReviewDataForCreateProto.CoachId)
	if err != nil {
		logger.ErrorLogger.Printf("id must be uuid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be uuid"})
		c.Set("InvalidUpdate", struct{}{})
		return
	}

	//CoachId validate
	_, err = uuid.Parse(coachReviewDataForCreateProto.CoachId)
	if err != nil {
		logger.ErrorLogger.Printf("id must be uuid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be uuid"})
		return
	}

	//Body validate
	if len(coachReviewDataForCreateProto.Body) < 10 || len(coachReviewDataForCreateProto.Body) > 255 {
		logger.ErrorLogger.Printf("Review body must be between 10 and 255 characters long")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review body must be between 10 and 255 characters long"})
		return
	}

	//UserId validate
	_, err = uuid.Parse(coachReviewDataForCreateProto.UserId)
	if err != nil {
		logger.ErrorLogger.Printf("id must be uuid")
		c.JSON(http.StatusBadRequest, gin.H{"error": "id must be uuid"})
		return
	}

	createCoachReviewRequest.ReviewDataForCreate = coachReviewDataForCreateProto

	review, err := (*h.reviewClient).CreateCoachReview(context.TODO(), createCoachReviewRequest)
	if err != nil {
		logger.ErrorLogger.Printf("Failed CreateCoachReview: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	//add user!!!
	getUserByIdRequest := &userGRPC.GetUserByIdRequest{
		Id: review.ReviewObject.UserId,
	}

	user, err := (*h.userClient).GetUserById(context.Background(), getUserByIdRequest)
	if err != nil {
		logger.ErrorLogger.Printf("Failed GetUserById: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reviewWithUser := &coachGRPC.ReviewWithUser{
		ReviewObject: review.ReviewObject,
		UserObject:   user.UserObject,
	}

	c.JSON(http.StatusOK, gin.H{
		"reviewWithUser": reviewWithUser,
	})
}
