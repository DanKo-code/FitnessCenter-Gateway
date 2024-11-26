package rest

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
)

func RegisterHTTPEndpoints(router *gin.Engine, validator *validator.Validate, ssoClient *grpc.ClientConn) {
	h := NewHandler(ssoClient, validator)

	router.PUT("/users/:id", h.UpdateUser)
}
