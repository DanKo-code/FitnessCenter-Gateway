package server

import (
	_ "Gateway/docs"
	ssoRest "Gateway/internal/sso/delivery/rest"
	userRest "Gateway/internal/user/delivery/rest"
	logrusCustom "Gateway/pkg/logger"
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type App struct {
	httpServer *http.Server
	SSOClient  *grpc.ClientConn
	userClient *grpc.ClientConn
}

func NewApp() (*App, error) {

	logrusCustom.LogWithLocation(logrus.InfoLevel, fmt.Sprintf("Entered NewApp function"))

	conn, err := grpc.NewClient(os.Getenv("SSO_SERVICE_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("failed to connect to SSO server: %v", err))
		return nil, err
	}
	connUer, err := grpc.NewClient(os.Getenv("USER_SERVICE_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrusCustom.LogWithLocation(logrus.ErrorLevel, fmt.Sprintf("failed to connect to User server: %v", err))
		return nil, err
	}

	return &App{
		SSOClient:  conn,
		userClient: connUer,
	}, nil
}

func (app *App) Run(port string) error {

	logrusCustom.LogWithLocation(logrus.InfoLevel, fmt.Sprintf("Entered Run function"))

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"PUT"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//TODO add cors

	validate := validator.New()

	ssoRest.RegisterHTTPEndpoints(router, validate, app.SSOClient)

	userRest.RegisterHTTPEndpoints(router, validate, app.userClient)

	router.GET(os.Getenv("SWAGGER_PATH"), ginSwagger.WrapHandler(swaggerFiles.Handler))

	app.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := app.httpServer.ListenAndServe(); err != nil {
			logrusCustom.LogWithLocation(logrus.FatalLevel, fmt.Sprintf("Failed to listen and serve: %+v", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return app.httpServer.Shutdown(ctx)
}
