package logger

import (
	"go.uber.org/zap"
	"os"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

func init() {
	config := zap.NewProductionConfig()

	envLoglevel := os.Getenv("LOGLEVEL")
	if envLoglevel != "" {
		logLevel := zap.InfoLevel
		err := logLevel.Set(envLoglevel)
		if err != nil {
			panic(err)
		}
		config.Level.SetLevel(logLevel)
	}

	Logger, _ = config.Build()
	Sugar = Logger.Sugar()
}
