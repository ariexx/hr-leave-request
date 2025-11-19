package main

import (
	"fmt"
	"hr-leave-request/injector"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize app with wire
	app, err := injector.InitializeApp()
	if err != nil {
		logrus.Fatalf("failed to initialize app: %v", err)
	}

	// Start server
	port := 3000
	logrus.Infof("Starting HR Leave Request API on port %d", port)
	if err := app.Listen(fmt.Sprintf(":%d", port)); err != nil {
		logrus.Fatalf("failed to start server: %v", err)
	}
}
