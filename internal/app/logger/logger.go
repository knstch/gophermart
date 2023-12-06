package logger

import "go.uber.org/zap"

// Логер ошибки
func ErrorLogger(msg string, serverErr error) {
	var logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	defer logger.Sync()
	sugar.Errorf("Error: %v\nDetails: %v\n", msg, serverErr)
}

// Информативный логгер
func InfoLogger(msg string) {
	var logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	defer logger.Sync()
	sugar.Infof(msg)
}
