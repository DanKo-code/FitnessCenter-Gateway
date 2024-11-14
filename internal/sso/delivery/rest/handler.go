package rest

import (
	"Geteway/internal/sso/dtos"
	"Geteway/internal/sso/sso_errors"
	logrusCustom "Geteway/pkg/logger"
	"context"
	"fmt"
	ssoGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.sso"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Handler struct {
	ssoClient *grpc.ClientConn
}

func NewHandler(ssoClient *grpc.ClientConn) *Handler {
	return &Handler{
		ssoClient: ssoClient,
	}
}

func (h *Handler) SignUp(c *gin.Context) {
	suReq := &dtos.SignUpRequest{}

	if err := c.ShouldBindJSON(&suReq); err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error binding SignUpRequest: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fingerPrintValue, exists := c.Get(os.Getenv("APP_FINGERPRINT_REQUEST_KEY"))
	if !exists {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error binding SignUpRequest: %v", sso_errors.FingerPrintNotFoundInContext))
		c.JSON(http.StatusInternalServerError, gin.H{"error": sso_errors.FingerPrintNotFoundInContext})
		return
	}

	FingerPrintValueCasted, ok := fingerPrintValue.(string)
	if !ok {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error casting fingerprint to string: %v", sso_errors.FingerPrintNotFoundInContext))
		c.JSON(http.StatusInternalServerError, gin.H{"error": sso_errors.FingerPrintNotFoundInContext})
		return
	}

	suReq.FingerPrint = FingerPrintValueCasted

	//TODO add validation

	ssoClient := ssoGRPC.NewSSOClient(h.ssoClient)

	signUpRequest := &ssoGRPC.SignUpRequest{}
	signUpRequest.Name = suReq.Name
	signUpRequest.Email = suReq.Email
	signUpRequest.Password = suReq.Password
	signUpRequest.FingerPrint = suReq.FingerPrint

	upRes, err := ssoClient.SignUp(context.Background(), signUpRequest)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error SignUp: %v", err))

		//TODO add statuses?
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rteInt, err := strconv.Atoi(upRes.RefreshTokenExpiration)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error convert RefreshTokenExpiration to int: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refreshToken",
		upRes.RefreshToken,
		rteInt,
		"",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"accessToken":           upRes.GetAccessToken(),
		"accessTokenExpiration": upRes.GetAccessTokenExpiration(),
	})
}

func (h *Handler) SignIn(c *gin.Context) {
	siReq := &dtos.SignInRequest{}

	if err := c.ShouldBindJSON(siReq); err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error parsing SignInRequest: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fingerPrintValue, exists := c.Get(os.Getenv("APP_FINGERPRINT_REQUEST_KEY"))
	if !exists {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error binding SignUpRequest: %v", sso_errors.FingerPrintNotFoundInContext))
		c.JSON(http.StatusInternalServerError, gin.H{"error": sso_errors.FingerPrintNotFoundInContext})
		return
	}

	FingerPrintValueCasted, ok := fingerPrintValue.(string)
	if !ok {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error casting fingerprint to string: %v", sso_errors.FingerPrintNotFoundInContext))
		c.JSON(http.StatusInternalServerError, gin.H{"error": sso_errors.FingerPrintNotFoundInContext})
		return
	}

	siReq.FingerPrint = FingerPrintValueCasted

	//TODO add validation

	ssoClient := ssoGRPC.NewSSOClient(h.ssoClient)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	signIpRequest := &ssoGRPC.SignInRequest{
		Email:       siReq.Email,
		Password:    siReq.Password,
		FingerPrint: siReq.FingerPrint,
	}

	siRes, err := ssoClient.SignIn(ctx, signIpRequest)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error SignIp: %v", err))

		//TODO add statuses?
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rteInt, err := strconv.Atoi(siRes.GetRefreshTokenExpiration())
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error convert RefreshTokenExpiration to int: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refreshToken",
		siRes.RefreshToken,
		rteInt,
		"",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"accessToken":           siRes.GetAccessTokenExpiration(),
		"accessTokenExpiration": siRes.GetRefreshTokenExpiration(),
	})
}

func (h *Handler) LogOut(c *gin.Context) {

	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error getting refreshToken from cookie: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ssoClient := ssoGRPC.NewSSOClient(h.ssoClient)

	logOutRequest := &ssoGRPC.LogOutRequest{}
	logOutRequest.RefreshToken = refreshToken

	_, err = ssoClient.LogOut(context.Background(), logOutRequest)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error LogOut: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("refreshToken", "", -1, "/", "", false, true)

	c.Status(http.StatusOK)
}

func (h *Handler) Refresh(c *gin.Context) {
	fingerPrintValue, exists := c.Get(os.Getenv("APP_FINGERPRINT_REQUEST_KEY"))
	if !exists {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error binding SignUpRequest: %v", sso_errors.FingerPrintNotFoundInContext))
		c.JSON(http.StatusInternalServerError, gin.H{"error": sso_errors.FingerPrintNotFoundInContext})
		return
	}

	FingerPrintValueCasted, ok := fingerPrintValue.(string)
	if !ok {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error casting fingerprint to string: %v", sso_errors.FingerPrintNotFoundInContext))
		c.JSON(http.StatusInternalServerError, gin.H{"error": sso_errors.FingerPrintNotFoundInContext})
		return
	}

	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error getting refreshToken from cookie: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshRequest := &ssoGRPC.RefreshRequest{
		FingerPrint:  FingerPrintValueCasted,
		RefreshToken: refreshToken,
	}

	ssoClient := ssoGRPC.NewSSOClient(h.ssoClient)

	rRes, err := ssoClient.Refresh(context.Background(), refreshRequest)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error Refresh: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rteInt, err := strconv.Atoi(rRes.GetRefreshTokenExpiration())
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error convert RefreshTokenExpiration to int: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie(
		"refreshToken",
		rRes.GetRefreshToken(),
		rteInt,
		"",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"user":                  rRes.GetUser(),
		"accessToken":           rRes.GetAccessToken(),
		"accessTokenExpiration": rRes.GetAccessTokenExpiration(),
	})
}
