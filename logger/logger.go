package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func Init(mode string) *zap.Logger {
	var err error
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	if mode == "debug" {
		option := zap.AddStacktrace(zap.DPanicLevel)
		Logger, err = config.Build(option)
		if err != nil {
			log.Fatal("error creating logger : ", err.Error())
		}

		Logger.Debug("Logger started", zap.String("mode", "debug"))
	} else {
		option := zap.AddStacktrace(zap.DPanicLevel)
		Logger, err = config.Build(option)
		if err != nil {
			log.Fatal("error creating logger : ", err.Error())
		}

		Logger.Info("Logger started", zap.String("mode", "production"))
	}

	return Logger
}
