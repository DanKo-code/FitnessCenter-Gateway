package rest

import (
	"Gateway/internal/sso/dtos"
	"Gateway/internal/sso/sso_errors"
	logrusCustom "Gateway/pkg/logger"
	"context"
	"fmt"
	ssoGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.sso"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"net/http"
	"os"
	"strconv"
)

var (
	expirationDivider = 1000000000
)

type Handler struct {
	ssoClient *grpc.ClientConn
	validator *validator.Validate
}

func NewHandler(ssoClient *grpc.ClientConn, validator *validator.Validate) *Handler {
	return &Handler{
		ssoClient: ssoClient,
		validator: validator,
	}
}

// SignUp
// @Summary Sign-up a new user
// @Description Sign-up a new user with provided details including name, email, and password. The API also manages fingerprint for enhanced tracking.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param signUpRequest body dtos.SignUpRequestWithOutFingerPrint true "Details for the user sign-up"
// @Success 200 {object} dtos.SignUpResponse "Successful registration response containing access token, expiration, and user details"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /sso/signUp [post]
func (h *Handler) SignUp(c *gin.Context) {
	suReq := &dtos.SignUpRequest{}

	if err := c.ShouldBindJSON(&suReq); err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error binding SignUpRequest: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.validator.Struct(suReq)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error validating SignUpRequest: %v", err))

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fingerPrintValue, exists := c.Get(os.Getenv("APP_FINGERPRINT_REQUEST_KEY"))
	if !exists {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error getting fingerprint: %v", sso_errors.FingerPrintNotFoundInContext))
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

	ateInt, err := strconv.Atoi(upRes.AccessTokenExpiration)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error convert AccessTokenExpiration to int: %v", err))
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
		rteInt/expirationDivider,
		"",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"accessToken":           upRes.GetAccessToken(),
		"accessTokenExpiration": ateInt / expirationDivider,
		"user":                  upRes.GetUser(),
	})
}

// SignIn
// @Summary Sign-in  user
// @Description Sign-in for existing user with provided details including email, and password. The API also manages fingerprint for enhanced tracking.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param signInRequest body dtos.SignInRequestWithoutFingerprint true "Details for the user sign-in"
// @Success 200 {object} dtos.SignInResponse "Successful sign-in response containing access token, expiration, and user details"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /sso/signIn [post]
func (h *Handler) SignIn(c *gin.Context) {
	siReq := &dtos.SignInRequest{}

	if err := c.ShouldBindJSON(siReq); err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error parsing SignInRequest: %v", err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.validator.Struct(siReq)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error validating SignInRequest: %v", err))

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

	ssoClient := ssoGRPC.NewSSOClient(h.ssoClient)

	signIpRequest := &ssoGRPC.SignInRequest{
		Email:       siReq.Email,
		Password:    siReq.Password,
		FingerPrint: siReq.FingerPrint,
	}

	siRes, err := ssoClient.SignIn(context.Background(), signIpRequest)
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error SignIp: %v", err))

		//TODO add statuses?
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ateInt, err := strconv.Atoi(siRes.GetAccessTokenExpiration())
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error convert AccessTokenExpiration to int: %v", err))
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
		siRes.GetRefreshToken(),
		rteInt/expirationDivider,
		"",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"accessToken":           siRes.GetAccessToken(),
		"accessTokenExpiration": ateInt / expirationDivider,
	})
}

// LogOut
// @Summary Log-out  user
// @Description Log-out for entered user. The API also manages fingerprint for enhanced tracking.
// @Tags Authentication
// @Accept json
// @Success 200 "Successful log-out"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /sso/logOut [post]
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

// Refresh
// @Summary Refreshing tokens for accessing secured resources
// @Description Refreshing tokens for accessing secured resources. Getting Refresh token from cookies. The API also manages fingerprint for enhanced tracking.
// @Tags Authentication
// @Accept json
// @Success 200 {object} dtos.RefreshResponse "Successful Refresh tokens response containing access token, expiration, and user details"
// @Failure 400 {object} map[string]string "Invalid request or validation error"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /sso/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	fingerPrintValue, exists := c.Get(os.Getenv("APP_FINGERPRINT_REQUEST_KEY"))
	if !exists {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("Error getting fingerprint	: %v", sso_errors.FingerPrintNotFoundInContext))
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
