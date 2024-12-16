package rest

import (
	"Gateway/pkg/logger"
	"context"
	"fmt"
	abonementGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.abonement"
	"github.com/gin-gonic/gin"
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
