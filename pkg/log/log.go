package log

import (
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Environment string

const (
	Production  Environment = "production"
	Development             = "development"
)

var Z *zap.Logger
var S *zap.SugaredLogger

func InitializeLogger(env Environment, logLevel zapcore.Level) {
	var loggerErr error
	if env == Production {
		Z, loggerErr = zap.NewProduction(zap.IncreaseLevel(logLevel))
	} else {
		Z, loggerErr = zap.NewDevelopment(zap.IncreaseLevel(logLevel))
	}

	if loggerErr != nil {
		log.Fatal("Failed to initialize zap logger: ", loggerErr)
	}

	defer Z.Sync()
	S = Z.Sugar()
}
