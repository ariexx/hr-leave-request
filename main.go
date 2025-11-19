package main

import (
	"fmt"
	"hr-leave-request/config"
	"hr-leave-request/injector"

	"github.com/sirupsen/logrus"
)

func main() {
	// load configuration and initialize dependencies using wire
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("failed to load config: %v", err)
	}

	// Initialize app with wire
	app, err := injector.InitializeApp()
	if err != nil {
		logrus.Fatalf("failed to initialize app: %v", err)
	}

	// Start server
	port := cfg.AppConfig.Port
	logrus.Infof("Starting HR Leave Request API on port %d", port)
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		logrus.Fatalf("failed to start server: %v", err)
	}
}
