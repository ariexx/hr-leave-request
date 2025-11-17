package main

import (
	"hr-leave-request/config"

	"github.com/sirupsen/logrus"
)

func main() {
	// load config
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("failed to load config: %v", err)
	}

	// setup logger based on environment
	if cfg.AppConfig.Environment == "production" {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logrus.Infof("Starting application in %s mode on port %d", cfg.AppConfig.Environment, cfg.AppConfig.Port)
}
