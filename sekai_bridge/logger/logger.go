package logger

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Init(mode string) *zap.Logger {
	var logger *zap.Logger
	var err error
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	if mode == "debug" {
		option := zap.AddStacktrace(zap.DPanicLevel)
		logger, err = config.Build(option)
		if err != nil {
			log.Fatal("error creating logger : ", err.Error())
		}

		logger.Debug("Logger started", zap.String("mode", "debug"))
	} else {
		option := zap.AddStacktrace(zap.DPanicLevel)
		logger, err = config.Build(option)
		if err != nil {
			log.Fatal("error creating logger : ", err.Error())
		}

		logger.Info("Logger started", zap.String("mode", "production"))
	}

	return logger
}
