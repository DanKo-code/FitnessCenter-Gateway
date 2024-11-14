package main

import (
	"Geteway/internal/server"
	logrusCustom "Geteway/pkg/logger"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	logrusCustom.InitLogger()

	err := godotenv.Load()
	if err != nil {
		logrusCustom.LogWithLocation(logrus.FatalLevel, fmt.Sprintf("Error loading .env file: %s", err))
	}

	logrusCustom.LogWithLocation(logrus.InfoLevel, "Successfully loaded environment variables")

	appGRPC, err := server.NewApp()
	if err != nil {
		logrusCustom.LogWithLocation(logrus.FatalLevel, fmt.Sprintf("Error initializing app: %s", err))
	}

	err = appGRPC.Run(os.Getenv("APP_PORT"))
	if err != nil {
		logrusCustom.LogWithLocation(logrus.FatalLevel, "Error running server")
	}
}
