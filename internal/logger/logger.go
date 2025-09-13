package logger

import "go.uber.org/zap"

var L *zap.Logger

func InitGlobal() {
	var err error
	L, err = NewLogger()
	if err != nil {
		panic(err)
	}
}

func NewLogger() (*zap.Logger, error) {
	//return zap.NewDevelopment()
	return zap.NewProduction()
}
