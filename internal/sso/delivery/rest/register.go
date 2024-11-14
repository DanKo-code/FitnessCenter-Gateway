package rest

import (
	"Geteway/internal/sso/delivery/rest/middlewares"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func RegisterHTTPEndpoints(router *gin.Engine, ssoClient *grpc.ClientConn) {
	h := NewHandler(ssoClient)

	ssoWithFingerprint := router.Group("/sso")
	{
		ssoWithFingerprint.Use(middlewares.FingerprintMiddleware())
		ssoWithFingerprint.POST("/signUp", h.SignUp)
		ssoWithFingerprint.POST("/signIn", h.SignIn)
		ssoWithFingerprint.POST("/refresh", h.Refresh)
	}

	ssoWithoutFingerprint := router.Group("/sso")
	{
		ssoWithoutFingerprint.POST("/logOut", h.LogOut)
	}
}
