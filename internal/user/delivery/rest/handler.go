package rest

import (
	"Gateway/internal/user/dtos"
	logger "Gateway/pkg/logger"
	"context"
	"fmt"
	userGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.user"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"io"
	"net/http"
)

type Handler struct {
	userClient *grpc.ClientConn
	validator  *validator.Validate
}

func NewHandler(ssoClient *grpc.ClientConn, validator *validator.Validate) *Handler {
	return &Handler{
		userClient: ssoClient,
		validator:  validator,
	}
}

func (h *Handler) UpdateUser(c *gin.Context) {

	userId := c.Param("id")

	cmd := &dtos.User{}

	form, err := c.MultipartForm()
	if err != nil {
		return
	}

	name := form.Value["name"]
	email := form.Value["email"]
	role := form.Value["role"]
	photo := form.File["photo"]

	err = h.validator.Struct(cmd)
	if err != nil {
		logger.ErrorLogger.Printf("Error validating UpdateUserRequest: %v", err)

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userClient := userGRPC.NewUserClient(h.userClient)

	stream, err := userClient.UpdateUser(context.Background())
	if err != nil {
		fmt.Printf("failed to stat file: %v\n", err)
	}

	userDataForUpdate := &userGRPC.UserDataForUpdate{
		Id:    userId,
		Email: email[0],
		Name:  name[0],
		Role:  role[0],
	}

	updateUserRequestUserDataForUpdate := &userGRPC.UpdateUserRequest_UserDataForUpdate{
		UserDataForUpdate: userDataForUpdate,
	}

	createUserRequest := &userGRPC.UpdateUserRequest{
		Payload: updateUserRequestUserDataForUpdate,
	}

	err = stream.Send(createUserRequest)
	if err != nil {
		return
	}

	if len(photo) > 0 {
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

			err = stream.Send(&userGRPC.UpdateUserRequest{
				Payload: &userGRPC.UpdateUserRequest_UserPhoto{
					UserPhoto: buffer[:n],
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
		"user": res,
	})
}
