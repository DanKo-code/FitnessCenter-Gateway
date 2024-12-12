package server

import (
	_ "Gateway/docs"
	abonementRest "Gateway/internal/abonement/rest"
	ssoRest "Gateway/internal/sso/delivery/rest"
	userRest "Gateway/internal/user/delivery/rest"
	logger "Gateway/pkg/logger"
	"context"
	abonementGRPC "github.com/DanKo-code/FitnessCenter-Protobuf/gen/FitnessCenter.protobuf.abonement"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
	httpServer      *http.Server
	SSOClient       *grpc.ClientConn
	userClient      *grpc.ClientConn
	abonementClient *abonementGRPC.AbonementClient
}

func NewApp() (*App, error) {

	conn, err := grpc.NewClient(os.Getenv("SSO_SERVICE_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.ErrorLogger.Printf("failed to connect to SSO server: %v", err)
		return nil, err
	}
	connUer, err := grpc.NewClient(os.Getenv("USER_SERVICE_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.ErrorLogger.Printf("failed to connect to User server: %v", err)
		return nil, err
	}
	connAbonement, err := grpc.NewClient(os.Getenv("ABONEMENT_SERVICE_PORT"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.ErrorLogger.Printf("failed to connect to User server: %v", err)
		return nil, err
	}
	abonementClient := abonementGRPC.NewAbonementClient(connAbonement)

	return &App{
		SSOClient:       conn,
		userClient:      connUer,
		abonementClient: &abonementClient,
	}, nil
}

func (app *App) Run(port string) error {

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3333"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	//TODO add cors

	validate := validator.New()

	ssoRest.RegisterHTTPEndpoints(router, validate, app.SSOClient)

	userRest.RegisterHTTPEndpoints(router, validate, app.userClient)

	abonementRest.RegisterHTTPEndpoints(router, app.abonementClient)

	router.GET(os.Getenv("SWAGGER_PATH"), ginSwagger.WrapHandler(swaggerFiles.Handler))

	app.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := app.httpServer.ListenAndServe(); err != nil {
			logger.FatalLogger.Printf("Failed to listen and serve: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Interrupt)

	<-quit

	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdown()

	return app.httpServer.Shutdown(ctx)
}
