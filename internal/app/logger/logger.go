// Package logger provides logging functions.
package logger

import "go.uber.org/zap"

// An error logger accepting a message as a string and error.
func ErrorLogger(msg string, serverErr error) {
	var logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	defer logger.Sync()
	sugar.Errorf("Error: %v\nDetails: %v\n", msg, serverErr)
}

// An info logger accepting message.
func InfoLogger(msg string) {
	var logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	var sugar = *logger.Sugar()
	defer logger.Sync()
	sugar.Infof(msg)
}
