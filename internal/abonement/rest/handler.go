package rest

import (
	"Gateway/pkg/logger"
	"context"
	"fmt"
	abonementGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.abonement"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net/http"
	"strconv"
	"strings"
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

		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"abonements": abonements,
	})
}

func (h *Handler) CreateAbonement(c *gin.Context) {

	form, err := c.MultipartForm()
	if err != nil {
		return
	}

	title := form.Value["title"]
	validityPeriod := form.Value["validity_period"]
	visitingTime := form.Value["visiting_time"]
	price := form.Value["price"]
	parsePrice, err := strconv.ParseInt(price[0], 10, 32)
	if err != nil {
		logger.ErrorLogger.Printf("Failed ParseInt: %s", err)
		return
	}
	services := form.Value["services"]
	photo := form.File["photo"]

	if services == nil {
		return
	}

	if len(services) == 0 {
		return
	}

	servicesIds := strings.Split(services[0], ",")

	stream, err := (*h.abonementClient).CreateAbonement(context.Background())
	if err != nil {
		fmt.Printf("failed to stat file: %v\n", err)
	}

	abonementDataForCreate := &abonementGRPC.AbonementDataForCreate{
		Title:        title[0],
		Validity:     validityPeriod[0],
		VisitingTime: visitingTime[0],
		Price:        int32(parsePrice),
		ServicesIds:  servicesIds,
	}

	createAbonementRequestAbonementDataForCreate := &abonementGRPC.CreateAbonementRequest_AbonementDataForCreate{
		AbonementDataForCreate: abonementDataForCreate,
	}

	createAbonementRequest := &abonementGRPC.CreateAbonementRequest{
		Payload: createAbonementRequestAbonementDataForCreate,
	}

	err = stream.Send(createAbonementRequest)
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

			err = stream.Send(&abonementGRPC.CreateAbonementRequest{
				Payload: &abonementGRPC.CreateAbonementRequest_AbonementPhoto{
					AbonementPhoto: buffer[:n],
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
		"abonement": res.GetAbonementWithServices(),
	})
}

func (h *Handler) UpdateAbonement(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		return
	}

	id := form.Value["id"]
	title := form.Value["title"]
	validityPeriod := form.Value["validity_period"]
	visitingTime := form.Value["visiting_time"]
	price := form.Value["price"]
	parsePrice, err := strconv.ParseInt(price[0], 10, 32)
	if err != nil {
		logger.ErrorLogger.Printf("Failed ParseInt: %s", err)
		return
	}
	services := form.Value["services"]
	photo := form.File["photo"]

	if services == nil {
		return
	}

	if len(services) == 0 {
		return
	}

	servicesIds := strings.Split(services[0], ",")

	stream, err := (*h.abonementClient).UpdateAbonement(context.Background())
	if err != nil {
		fmt.Printf("failed to stat file: %v\n", err)
	}

	abonementDataForUpdate := &abonementGRPC.AbonementDataForUpdate{
		Id:           id[0],
		Title:        title[0],
		Validity:     validityPeriod[0],
		VisitingTime: visitingTime[0],
		Price:        int32(parsePrice),
		ServicesIds:  servicesIds,
	}

	updateAbonementRequestAbonementDataForUpdate := &abonementGRPC.UpdateAbonementRequest_AbonementDataForUpdate{
		AbonementDataForUpdate: abonementDataForUpdate,
	}

	updateAbonementRequest := &abonementGRPC.UpdateAbonementRequest{
		Payload: updateAbonementRequestAbonementDataForUpdate,
	}

	err = stream.Send(updateAbonementRequest)
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

			err = stream.Send(&abonementGRPC.UpdateAbonementRequest{
				Payload: &abonementGRPC.UpdateAbonementRequest_AbonementPhoto{
					AbonementPhoto: buffer[:n],
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
		"abonement": res.GetAbonementWithServices(),
	})
}

func (h *Handler) DeleteAbonement(c *gin.Context) {

	id := c.Param("id")

	convertedId, err := uuid.Parse(id)
	if err != nil {

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("invalid id format")})
		return
	}

	_ = convertedId

	deleteAbonementByIdRequest := &abonementGRPC.DeleteAbonementByIdRequest{
		Id: id,
	}

	deletedAbonementRes, err := (*h.abonementClient).DeleteAbonementById(context.TODO(), deleteAbonementByIdRequest)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"abonement": deletedAbonementRes.GetAbonementObject(),
	})
}
