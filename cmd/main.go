package main

import (
	"os"
	"runtime/debug"

	"github.com/4udiwe/comments-feed/internal/app"
	"github.com/sirupsen/logrus"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("panic in main: %v\n%s", r, debug.Stack())
		}
	}()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		logrus.Warn("empty configpath env")
		configPath = "config/config.yaml"
	}

	application := app.New(configPath)
	application.Start()
}
