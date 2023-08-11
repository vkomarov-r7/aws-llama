package log

import "go.uber.org/zap"

var Logger *zap.SugaredLogger

func InitLogging() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	Logger = logger.Sugar()
}
