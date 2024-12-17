package rest

import (
	"context"
	"fmt"
	coachGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.coach"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net/http"
	"strings"
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
	coaches, err := (*h.coachClient).GetCoachesWithServicesWithReviewsWithUsers(c.Request.Context(), &emptypb.Empty{})
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

func (h *Handler) CreateCoach(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		return
	}

	name := form.Value["name"]
	description := form.Value["description"]
	services := form.Value["services"]
	photo := form.File["photo"]

	if services == nil {
		return
	}

	if len(services) == 0 {
		return
	}

	servicesIds := strings.Split(services[0], ",")

	stream, err := (*h.coachClient).CreateCoach(context.Background())
	if err != nil {
		fmt.Printf("failed to stat file: %v\n", err)
	}

	coachDataForCreate := &coachGRPC.CoachDataForCreate{
		Name:            name[0],
		Description:     description[0],
		CoachServiceIds: servicesIds,
	}

	createCoachRequestCoachDataForCreate := &coachGRPC.CreateCoachRequest_CoachDataForCreate{
		CoachDataForCreate: coachDataForCreate,
	}

	createCoachRequest := &coachGRPC.CreateCoachRequest{
		Payload: createCoachRequestCoachDataForCreate,
	}

	err = stream.Send(createCoachRequest)
	if err != nil {
		return
	}

	if photo != nil && len(photo) > 0 {
		buffer := make([]byte, 1024*1024)
		file, err := photo[0].Open()
		if err != nil {
			return
		}

		for {
			n, err := file.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}

			err = stream.Send(&coachGRPC.CreateCoachRequest{
				Payload: &coachGRPC.CreateCoachRequest_CoachPhoto{
					CoachPhoto: buffer[:n],
				},
			},
			)
			if err != nil {
				return
			}
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coach": res.GetCoachWithServices(),
	})
}

func (h *Handler) UpdateCoach(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		return
	}

	id := form.Value["id"]
	name := form.Value["name"]
	description := form.Value["description"]
	services := form.Value["services"]
	photo := form.File["photo"]

	if services == nil {
		return
	}

	if len(services) == 0 {
		return
	}

	servicesIds := strings.Split(services[0], ",")

	stream, err := (*h.coachClient).UpdateCoach(context.Background())
	if err != nil {
		fmt.Printf("failed to stat file: %v\n", err)
	}

	coachDataForUpdate := &coachGRPC.CoachDataForUpdate{
		Id:              id[0],
		Name:            name[0],
		Description:     description[0],
		CoachServiceIds: servicesIds,
	}

	updateCoachRequestCoachDataForUpdate := &coachGRPC.UpdateCoachRequest_CoachDataForUpdate{
		CoachDataForUpdate: coachDataForUpdate,
	}

	updateCoachRequest := &coachGRPC.UpdateCoachRequest{
		Payload: updateCoachRequestCoachDataForUpdate,
	}

	err = stream.Send(updateCoachRequest)
	if err != nil {
		return
	}

	if photo != nil && len(photo) > 0 {
		buffer := make([]byte, 1024*1024)
		file, err := photo[0].Open()
		if err != nil {
			return
		}

		for {
			n, err := file.Read(buffer)
			if err == io.EOF {
				break
			}
			if err != nil {
				return
			}

			err = stream.Send(&coachGRPC.UpdateCoachRequest{
				Payload: &coachGRPC.UpdateCoachRequest_CoachPhoto{
					CoachPhoto: buffer[:n],
				},
			},
			)
			if err != nil {
				return
			}
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coach": res.GetCoachWithServices(),
	})
}

func (h *Handler) DeleteCoach(c *gin.Context) {
	id := c.Param("id")

	convertedId, err := uuid.Parse(id)
	if err != nil {

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid id format")})
		return
	}

	_ = convertedId

	deleteCoachByIdRequest := &coachGRPC.DeleteCoachByIdRequest{
		Id: id,
	}

	deletedCoachRes, err := (*h.coachClient).DeleteCoachById(context.TODO(), deleteCoachByIdRequest)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"coach": deletedCoachRes.GetCoachObject(),
	})
}
